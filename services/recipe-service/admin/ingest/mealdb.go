package ingest

import "fmt"

type MealDBResponse struct {
	Meals []MapMeal `json:"meals"`
}

type MapMeal map[string]any

type MealIngredient struct {
	Name        string
	Measurement string
}

func (m MapMeal) ToIngredientList() []MealIngredient {
	var ingredients []MealIngredient

	for i := 1; i <= 20; i++ {
		name, _ := m[fmt.Sprintf("strIngredient%d", i)].(string)
		measure, _ := m[fmt.Sprintf("strMeasure%d", i)].(string)

		if name != "" && name != "null" {
			ingredients = append(ingredients, MealIngredient{
				Name:        name,
				Measurement: measure,
			})
		}
	}

	return ingredients
}
