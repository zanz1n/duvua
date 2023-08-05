use serenity::model::prelude::{
    application_command::CommandDataOption, command::CommandOptionType,
};

#[inline]
pub fn get_sub_command(options: &Vec<CommandDataOption>) -> Option<CommandDataOption> {
    for option in options.iter() {
        if option.kind == CommandOptionType::SubCommand {
            return Some(option.clone());
        }
    }

    None
}

#[inline]
pub fn get_option<T: ToString>(
    options: &Vec<CommandDataOption>,
    name: T,
) -> Option<CommandDataOption> {
    for option in options.iter() {
        if option.name == name.to_string() {
            return Some(option.clone());
        }
    }

    None
}
