use std::fmt::{self, Display, Formatter};
use std::time::Duration;

use futures::channel::oneshot::{channel, Sender};
use futures::future::FutureExt;
use futures::select;
use tokio::sync::mpsc::{unbounded_channel, UnboundedSender};

pub struct IOStat {
    collector: UnboundedSender<Duration>,
    stator: UnboundedSender<Sender<Box<Option<StatResult>>>>,
    shutdown: Option<Sender<()>>,
}

impl IOStat {
    pub fn start() -> IOStat {
        let (collector_tx, mut collector_rx) = unbounded_channel();
        let (stator_tx, mut stator_rx) = unbounded_channel::<Sender<_>>();
        let (shutdown_tx, shutdown_rx) = channel();

        let _ = tokio::spawn(async move {
            let mut durations = Vec::new();
            let mut fused_shutdown = shutdown_rx.fuse();
            loop {
                select! {
                    duration = collector_rx.recv().fuse() => {
                        if let Some(duration) = duration {
                            durations.push(duration);
                        } else {
                            break;
                        }
                    }
                    stator = stator_rx.recv().fuse() => {
                        if let Some(stator) = stator {
                            let _ = stator.send(Box::new(StatResult::statistic(&mut durations)));
                        } else {
                            break;
                        }
                    }
                    _ = fused_shutdown => {
                        break;
                    }
                }
            }
        });

        IOStat {
            collector: collector_tx,
            stator: stator_tx,
            shutdown: Some(shutdown_tx),
        }
    }

    pub fn collect(&self, dur: Duration) {
        let _ = self.collector.send(dur).expect("collect durations");
    }

    pub async fn stat(&self) -> Box<Option<StatResult>> {
        let (tx, rx) = channel();
        let _ = self.stator.send(tx);
        rx.await.expect("stat result")
    }
}

impl Drop for IOStat {
    fn drop(&mut self) {
        if let Some(shutdown) = self.shutdown.take() {
            let _ = shutdown.send(());
        }
    }
}

pub struct StatResult {
    percent_10: Duration,
    percent_25: Duration,
    percent_50: Duration,
    percent_75: Duration,
    percent_90: Duration,
    percent_95: Duration,
    percent_99: Duration,
}

impl StatResult {
    fn statistic(durations: &mut Vec<Duration>) -> Option<Self> {
        if durations.is_empty() {
            None
        } else {
            durations.sort();
            let len = durations.len();
            let result = Self {
                percent_10: durations[len / 10],
                percent_25: durations[len / 4],
                percent_50: durations[len / 2],
                percent_75: durations[len * 3 / 4],
                percent_90: durations[len * 9 / 10],
                percent_95: durations[len * 19 / 20],
                percent_99: durations[len * 99 / 100],
            };
            durations.clear();
            Some(result)
        }
    }
}

impl Display for StatResult {
    fn fmt(&self, f: &mut Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "\nIO Latency distribution:\n  10% in {:?}\n  25% in {:?}\n  50% in {:?}\n  75% in {:?}\n  90% in {:?}\n  95% in {:?}\n  99% in {:?}\n",
            self.percent_10,
            self.percent_25,
            self.percent_50,
            self.percent_75,
            self.percent_90,
            self.percent_95,
            self.percent_99
        )
    }
}
