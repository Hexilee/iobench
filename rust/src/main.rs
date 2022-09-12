use std::convert::Infallible;
use std::fs::{remove_file, File};
use std::io::Write;
use std::net::SocketAddr;
use std::path::Path;

use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Request, Response, Server, StatusCode};
use uuid::Uuid;

const DATA: [u8; 4096] = [b'0'; 4096];

async fn fast_handler(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    Ok(Response::new(Body::from(DATA.as_slice())))
}

async fn slow_handler(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    if let Err(err) = tokio::task::spawn_blocking(slow_task)
        .await
        .expect("join err")
    {
        return Ok(Response::builder()
            .status(StatusCode::INTERNAL_SERVER_ERROR)
            .body(Body::from(err.to_string()))
            .unwrap());
    }
    Ok(Response::new(Body::from(DATA.as_slice())))
}

fn slow_task() -> anyhow::Result<()> {
    let filename = Uuid::new_v4().to_string();
    let filepath = Path::new("./data").join(filename);
    let mut file = File::create(&filepath)?;
    file.write_all(DATA.as_slice())?;
    file.sync_all()?;
    drop(file);
    remove_file(&filepath)?;
    Ok(())
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
