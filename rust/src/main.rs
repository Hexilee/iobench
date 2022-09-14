use std::convert::Infallible;
use std::fs::{read, remove_file, File};
use std::io::Write;
use std::net::SocketAddr;
use std::path::Path;

use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Request, Response, Server, StatusCode};
use uuid::Uuid;

async fn handle_task<E: 'static + Send + ToString>(
    task: fn() -> Result<Vec<u8>, E>,
) -> Result<Response<Body>, Infallible> {
    match tokio::task::spawn_blocking(task).await.expect("join err") {
        Ok(data) => Ok(Response::new(Body::from(data))),
        Err(err) => Ok(Response::builder()
            .status(StatusCode::INTERNAL_SERVER_ERROR)
            .body(Body::from(err.to_string()))
            .unwrap()),
    }
}

async fn fast_handler(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    handle_task(|| read("./data/data.txt")).await
}

async fn slow_handler(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    handle_task(slow_task).await
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
