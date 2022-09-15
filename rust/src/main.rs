#![feature(async_closure)]

use std::convert::Infallible;
use std::fs::{read, remove_file, File};
use std::io::Write;
use std::net::SocketAddr;
use std::path::Path;
use std::sync::Arc;
use std::time::Instant;

use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Response, Server, StatusCode};
use once_cell::sync::Lazy;
use stat::IOStat;
use uuid::Uuid;

mod stat;

static FAST_STAT: Lazy<IOStat> = Lazy::new(|| IOStat::start());
static SLOW_STAT: Lazy<IOStat> = Lazy::new(|| IOStat::start());

async fn handle_task<E: 'static + Send + ToString>(
    task: fn() -> Result<Vec<u8>, E>,
    stat: &IOStat,
) -> Result<Response<Body>, Infallible> {
    let start = Instant::now();
    match tokio::task::spawn_blocking(task).await.expect("join err") {
        Ok(data) => {
            stat.collect(start.elapsed());
            Ok(Response::new(Body::from(data)))
        }
        Err(err) => Ok(Response::builder()
            .status(StatusCode::INTERNAL_SERVER_ERROR)
            .body(Body::from(err.to_string()))
            .unwrap()),
    }
}

async fn stat_task(stat: &IOStat) -> Result<Response<Body>, Infallible> {
    Ok(Response::new(Body::from(match *stat.stat().await {
        Some(stat) => stat.to_string(),
        None => String::new(),
    })))
}

fn slow_task() -> anyhow::Result<Vec<u8>> {
    let data = read("./data/data.txt")?;
    let filename = Uuid::new_v4().to_string();
    let filepath = Path::new("./data/tmp").join(filename);
    let mut file = File::create(&filepath)?;
    file.write_all(data.as_slice())?;
    file.sync_all()?;
    drop(file);
    remove_file(&filepath)?;
    Ok(data)
}

#[tokio::main]
async fn main() {
    let guard = Arc::new(
        pprof::ProfilerGuardBuilder::default()
            .frequency(1000)
            .build()
            .unwrap(),
    );
    let addr = SocketAddr::from(([0, 0, 0, 0], 8000));
    let make_svc = make_service_fn(|_conn| {
        let guard_ref = guard.clone();
        async move {
            Ok::<_, Infallible>(service_fn(move |req| {
                let guard_ref = guard_ref.clone();
                async move {
                    match req.uri().path() {
                        "/fast" => handle_task(|| read("./data/data.txt"), &*FAST_STAT).await,
                        "/slow" => handle_task(slow_task, &*SLOW_STAT).await,
                        "/stat/fast" => stat_task(&*FAST_STAT).await,
                        "/stat/slow" => stat_task(&*SLOW_STAT).await,
                        "/flamegraph.svg" => {
                            let mut graph = Vec::new();
                            if let Ok(report) = guard_ref.report().build() {
                                report.flamegraph(&mut graph).unwrap();
                            }
                            Ok(Response::new(Body::from(graph)))
                        }
                        _ => Ok(Response::builder()
                            .status(StatusCode::NOT_FOUND)
                            .body(Body::empty())
                            .unwrap()),
                    }
                }
            }))
        }
    });

    let server = Server::bind(&addr).serve(make_svc);

    // Run this server for... forever!
    if let Err(e) = server.await {
        eprintln!("server error: {}", e);
    }
}
