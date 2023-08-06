use super::hashmap_to_json_map;
use serde_json::Value;
use serenity::builder::{
    CreateActionRow as SerenityCreateActionRow, CreateButton, CreateComponents, CreateInputText,
    CreateSelectMenu,
};

#[derive(Clone, Debug)]
pub struct CreateActionRow(pub SerenityCreateActionRow);

impl CreateActionRow {
    pub fn add_button(&mut self, button: CreateButton) -> &mut Self {
        let components = self
            .0
             .0
            .entry("components")
            .or_insert_with(|| Value::from(Vec::<Value>::new()));
        let components_array = components.as_array_mut().expect("Must be an array");

        components_array.push(button.build());

        self
    }

    pub fn add_select_menu(&mut self, menu: CreateSelectMenu) -> &mut Self {
        let components = self
            .0
             .0
            .entry("components")
            .or_insert_with(|| Value::from(Vec::<Value>::new()));
        let components_array = components.as_array_mut().expect("Must be an array");

        components_array.push(menu.build());

        self
    }

    pub fn add_input_text(&mut self, input_text: CreateInputText) -> &mut Self {
        let components = self
            .0
             .0
            .entry("components")
            .or_insert_with(|| Value::from(Vec::<Value>::new()));
        let components_array = components.as_array_mut().expect("Must be an array");

        components_array.push(input_text.build());

        self
    }

    #[inline]
    pub fn to_components(&self) -> CreateComponents {
        self.to_owned().components()
    }

    #[inline]
    pub fn components(self) -> CreateComponents {
        CreateComponents(vec![hashmap_to_json_map(self.0 .0).into()])
    }
}

impl Default for CreateActionRow {
    fn default() -> Self {
        Self(SerenityCreateActionRow::default())
    }
}
