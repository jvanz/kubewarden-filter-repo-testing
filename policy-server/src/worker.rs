use crate::policies::Policy;
use crate::wasm::{EvalRequest, PolicyEvaluator};
use anyhow::Result;
use std::{collections::HashMap, thread, vec::Vec};
use tokio::sync::mpsc::{channel, Receiver, Sender};

pub(crate) struct Worker {
  evaluators: HashMap<String, PolicyEvaluator>,
  channel_rx: Receiver<EvalRequest>,
}

impl Worker {
  pub(crate) fn new(
    rx: Receiver<EvalRequest>,
    policies: HashMap<String, Policy>,
  ) -> Result<Worker> {
    let mut evs: HashMap<String, PolicyEvaluator> = HashMap::new();

    for (id, policy) in policies.iter() {
      let settings = policy.settings();

      let policy_evaluator = PolicyEvaluator::new(policy.wasm_module_path.clone(), settings)?;
      evs.insert(id.to_string(), policy_evaluator);
    }

    Ok(Worker {
      evaluators: evs,
      channel_rx: rx,
    })
  }

  pub(crate) fn run(mut self) {
    while let Some(req) = self.channel_rx.blocking_recv() {
      //TODO: handle error
      match self.evaluators.get_mut(&req.policy_id) {
        Some(policy_evaluator) => {
          let resp = policy_evaluator.validate(req.req);
          let _ = req.resp_chan.send(Some(resp));
        }
        None => {
          let _ = req.resp_chan.send(None);
        }
      }
    }
  }
}

pub(crate) struct WorkerPool {
  pool_size: usize,
  worker_tx_chans: Vec<Sender<EvalRequest>>,
  api_rx: Receiver<EvalRequest>,
}

impl WorkerPool {
  pub(crate) fn new(
    size: usize,
    policies: HashMap<String, Policy>,
    rx: Receiver<EvalRequest>,
  ) -> Result<WorkerPool> {
    let mut tx_chans = Vec::<Sender<EvalRequest>>::new();

    for n in 1..=size {
      let (tx, rx) = channel::<EvalRequest>(32);
      tx_chans.push(tx);
      let ps = policies.clone();

      thread::spawn(move || {
        let worker = Worker::new(rx, ps).unwrap();

        //TODO: better logging
        println!("worker {} loop start", n);
        worker.run();
        println!("worker {} loop exit", n);
      });
    }

    Ok(WorkerPool {
      pool_size: size,
      worker_tx_chans: tx_chans,
      api_rx: rx,
    })
  }

  pub(crate) fn run(mut self) {
    let mut next_worker_id = 0;
    while let Some(req) = self.api_rx.blocking_recv() {
      let _ = self.worker_tx_chans[next_worker_id].blocking_send(req);
      next_worker_id += 1;
      if next_worker_id >= self.pool_size {
        next_worker_id = 0;
      }
    }

    //TODO: should we also `join` the children threads here?
  }
}
