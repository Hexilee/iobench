use std::convert::Infallible;
use std::fs::{read, remove_file, File};
use std::io::Write;
use std::net::SocketAddr;
use std::path::Path;
use std::time::Instant;

use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Request, Response, Server, StatusCode};
use once_cell::sync::Lazy;
use uuid::Uuid;

use stat::IOStat;

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

async fn fast_handler(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    handle_task(|| read("./data/data.txt"), &*FAST_STAT).await
}

async fn slow_handler(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    handle_task(slow_task, &*SLOW_STAT).await
}

async fn stat_fast_handler(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    stat_task(&*FAST_STAT).await
}

async fn stat_slow_handler(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    stat_task(&*SLOW_STAT).await
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
    let addr = SocketAddr::from(([0, 0, 0, 0], 8000));
    let make_svc = make_service_fn(|_conn| async {
        Ok::<_, Infallible>(service_fn(|req| async {
            match req.uri().path() {
                "/fast" => fast_handler(req).await,
                "/slow" => slow_handler(req).await,
                "/stat/fast" => stat_fast_handler(req).await,
                "/stat/slow" => stat_slow_handler(req).await,
                _ => Ok(Response::builder()
                    .status(StatusCode::NOT_FOUND)
                    .body(Body::empty())
                    .unwrap()),
            }
        }))
    });

    let server = Server::bind(&addr).serve(make_svc);

    // Run this server for... forever!
    if let Err(e) = server.await {
        eprintln!("server error: {}", e);
    }
}
