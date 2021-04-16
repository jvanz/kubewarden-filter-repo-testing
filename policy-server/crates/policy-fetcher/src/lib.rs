use anyhow::{anyhow, Result};
use std::boxed::Box;
use url::Url;

pub mod fetcher;
mod https;
mod local;
pub mod registry;
pub mod sources;

use crate::registry::config::DockerConfig;

use crate::fetcher::Fetcher;
use crate::https::Https;
use crate::local::Local;
use crate::registry::Registry;
use crate::sources::Sources;

// Helper function, takes the URL of the WASM module and allocates
// the right struct to interact with it
pub(crate) fn parse_wasm_url(
    url: &str,
    docker_config: Option<DockerConfig>,
    download_dir: &str,
) -> Result<Box<dyn Fetcher>> {
    // we have to use url::Url instead of hyper::Uri because the latter one can't
    // parse urls like file://
    let parsed_url: Url = match url::Url::parse(url) {
        Ok(u) => u,
        Err(e) => {
            return Err(anyhow!("Invalid WASI url: {}", e));
        }
    };

    match parsed_url.scheme() {
        "file" => Ok(Box::new(Local::new(parsed_url.path()))),
        "http" | "https" => Ok(Box::new(Https::new(url.parse::<Url>()?, download_dir)?)),
        "registry" => Ok(Box::new(Registry::new(
            parsed_url,
            docker_config,
            download_dir,
        )?)),
        _ => Err(anyhow!("unknown scheme: {}", parsed_url.scheme())),
    }
}

pub async fn fetch_wasm_module(
    url: &str,
    download_dir: &str,
    docker_config: Option<DockerConfig>,
    sources: &Sources,
) -> Result<String> {
    parse_wasm_url(&url, docker_config, download_dir)?
        .fetch(sources)
        .await
}
