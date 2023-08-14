use rand::seq::SliceRandom;

pub struct RandomStringProvider {
    vec: Vec<String>,
}

impl RandomStringProvider {
    pub fn kiss_gifs() -> Self {
        let raw_file = include_str!("../../../../assets/kiss-gifs.txt");
        let vec = raw_file.split("\n").map(String::from).collect();

        Self { vec }
    }

    pub fn slap_gifs() -> Self {
        let raw_file = include_str!("../../../../assets/slap-gifs.txt");
        let vec = raw_file.split("\n").map(String::from).collect();

        Self { vec }
    }

    pub fn get_choice(&self) -> Option<&str> {
        match self.vec.choose(&mut rand::thread_rng()) {
            Some(v) => Some(v),
            None => None,
        }
    }
}
