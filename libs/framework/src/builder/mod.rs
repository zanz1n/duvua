use serde_json::{Map, Value};
use std::{
    collections::HashMap,
    hash::{BuildHasher, Hash},
};

pub mod button_action_row;
pub mod interaction_response;

pub(crate) fn hashmap_to_json_map<H, T>(map: HashMap<T, Value, H>) -> Map<String, Value>
where
    H: BuildHasher,
    T: Eq + Hash + ToString,
{
    map.into_iter().map(|(k, v)| (k.to_string(), v)).collect()
}
