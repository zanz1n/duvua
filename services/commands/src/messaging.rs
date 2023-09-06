use redis::{Client, Commands, RedisError};
use serde::Serialize;
use serde_json::Value;
use std::{
    io::{Error, ErrorKind},
    sync::{Arc, Mutex},
    thread,
};
use tokio::sync::broadcast::{channel, error::RecvError, Receiver, Sender};

#[derive(Debug, Clone)]
pub struct Message {
    pub channel: String,
    pub payload: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct Payload {
    pub kind: u8,
    pub data: Value,
}

impl Payload {
    pub fn from_json(value: &Value) -> Result<Self, Error> {
        let data = value.get("data").ok_or(Error::new(
            ErrorKind::InvalidData,
            "Field 'data' not provided",
        ))?;

        let kind = value.get("kind").ok_or(Error::new(
            ErrorKind::InvalidData,
            "Field 'kind' not provided",
        ))?;

        let kind: u8 = kind
            .as_u64()
            .ok_or(Error::new(
                ErrorKind::InvalidData,
                "Field 'kind' must be a unsigned integer",
            ))?
            .try_into()
            .or(Err(Error::new(
                ErrorKind::InvalidData,
                "Field 'kind' must be a unsigned 8 bit integer",
            )))?;

        Ok(Self {
            kind,
            data: data.clone(),
        })
    }
}

pub fn extract_payload_data(payloads: Vec<Payload>) -> Vec<Value> {
    let mut data = Vec::with_capacity(payloads.len());

    for payload in payloads {
        data.push(payload.data)
    }

    data
}

pub struct SubClient {
    client: Arc<Client>,
    recv: Mutex<Receiver<Message>>,
    sender: Sender<Message>,
    sub_list: Vec<&'static str>,
}

pub fn spawn_daemon(client: Arc<SubClient>) {
    thread::spawn(move || client.listen_loop());
}

impl SubClient {
    pub fn connect(url: &str) -> Result<Self, RedisError> {
        let client = Client::open(url)?;
        let client = Arc::new(client);

        client.get_connection()?;

        Ok(Self::new(client))
    }

    pub fn subscribe(&mut self, channel: &'static str) {
        self.sub_list.push(channel);
    }

    pub fn set(&self, key: &str, value: String) -> Result<(), RedisError> {
        let mut conn = self.client.get_connection()?;
        conn.set(key, value)
    }

    pub fn new(client: Arc<Client>) -> Self {
        let (sender, recv) = channel(10);

        Self {
            sub_list: Vec::new(),
            recv: Mutex::new(recv),
            sender,
            client,
        }
    }

    pub async fn recv(&self) -> Result<Message, RecvError> {
        let mut lock = self.recv.lock().unwrap();
        let res = lock.recv().await;
        drop(lock);
        res
    }

    fn listen_loop(&self) -> ! {
        let mut conn = self.client.get_connection().unwrap();
        let mut conn = conn.as_pubsub();

        for sub in self.sub_list.iter() {
            conn.subscribe(*sub).unwrap();
        }

        loop {
            let message = conn.get_message().unwrap();
            let channel = message.get_channel_name().to_owned();
            let payload = message.get_payload().unwrap();

            let msg = Message { channel, payload };

            self.sender.send(msg).unwrap();
        }
    }
}
