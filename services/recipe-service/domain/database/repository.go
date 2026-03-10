package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/admin/ingest"
	"github.com/willh-simpson/pantry-wizard/services/recipe-service/domain/model"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
var PLACEHOLDER_AUTHOR_ID = "00000000-0000-0000-0000-000000000000"

func CreateFullRecipe(db *sql.DB, req model.CreateRecipeRequest) (string, error) {
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	recipeQuery, recipeArgs, err := psql.
		Insert("recipes").
		Columns("title", "description", "instructions", "author_id", "prep_time_min", "calories", "budget_tier").
		Values(req.Title, req.Description, req.Instructions, req.AuthorID, req.PrepTime, req.Calories, req.BudgetTier).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return "", fmt.Errorf("failed to build recipy query: %v", err)
	}

	var recipeID string

	err = tx.
		QueryRowContext(ctx, recipeQuery, recipeArgs...).
		Scan(&recipeID)

	for _, ingredient := range req.Ingredients {
		var ingredientID string

		ingredientQuery, ingredientArgs, err := psql.
			Insert("ingredients").
			Columns("name", "category").
			Values(ingredient.Name, ingredient.Category).
			Suffix("ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id").
			ToSql()
		if err != nil {
			return "", fmt.Errorf("failed to build ingredient query: %v", err)
		}

		err = tx.
			QueryRowContext(ctx, ingredientQuery, ingredientArgs...).
			Scan(&ingredientID)
		if err != nil {
			return "", fmt.Errorf("failed to upsert ingredient \"%s\": %v", ingredient.Name, err)
		}

		linkQuery, linkArgs, err := psql.
			Insert("recipe_ingredients").
			Columns("recipe_id", "ingredient_id", "amount", "unit").
			Values(recipeID, ingredientID, ingredient.Amount, ingredient.Unit).
			ToSql()
		if err != nil {
			return "", fmt.Errorf("failed to build link query for ingredient \"%s\": %v", ingredient.Name, err)
		}

		_, err = tx.ExecContext(ctx, linkQuery, linkArgs...)
		if err != nil {
			return "", fmt.Errorf("failed to link ingredient \"%s\": %v", ingredient.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %v", err)
	}

	return recipeID, nil
}

func SearchRecipes(db *sql.DB, title string, maxBudget int, maxPrepTime int) ([]model.Recipe, error) {
	queryBuilder := psql.
		Select("*").
		From("recipes")

	if title != "" {
		queryBuilder = queryBuilder.Where(squirrel.ILike{
			"title": fmt.Sprintf("%%%s%%", title),
		})
	}

	if maxBudget > 0 {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{
			"budget_tier": maxBudget,
		})
	}

	if maxPrepTime > 0 {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{
			"prep_time_min": maxPrepTime,
		})
	}

	queryBuilder = queryBuilder.OrderBy("created_at DESC")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %v", err)
	}
	defer rows.Close()

	var recipes []model.Recipe
	for rows.Next() {
		var r model.Recipe

		if err := rows.Scan(
			&r.ID,
			&r.Title,
			&r.TimesMadeGlobally,
			&r.Description,
			&r.Instructions,
			&r.AuthorID,
			&r.PrepTimeMinutes,
			&r.Calories,
			&r.BudgetTier,
			&r.ImageURL,
			&r.CreatedAt,
			&r.UpdatedAt,
		); err != nil {
			return nil, err
		}

		recipes = append(recipes, r)
	}

	return recipes, nil
}

func RecipeFromMealDB(db *sql.DB, ctx context.Context, meal ingest.MapMeal) (string, error) {
	title := meal["strMeal"].(string)
	instructions := meal["strInstructions"].(string)
	imageURL := meal["strMealThumb"].(string)
	ingredients := meal.ToIngredientList()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	ingredientIDs := make(map[string]string)
	for _, ingredient := range ingredients {
		var id string

		query, args, _ := psql.
			Insert("ingredients").
			Columns("name").
			Values(ingredient.Name).
			Suffix("ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id").
			ToSql()

		if err := tx.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
			return "", fmt.Errorf("ingredient upsert failed: %w", err)
		}

		ingredientIDs[ingredient.Name] = id
	}

	var recipeID string

	recipeQuery, recipeArgs, _ := psql.
		Insert("recipes").
		Columns("title", "instructions", "image_url", "author_id").
		Values(title, instructions, imageURL, PLACEHOLDER_AUTHOR_ID).
		Suffix("RETURNING id").
		ToSql()

	if err := tx.QueryRowContext(ctx, recipeQuery, recipeArgs...).Scan(&recipeID); err != nil {
		return "", fmt.Errorf("recipe insert failed: %w", err)
	}

	for _, ingredient := range ingredients {
		linkQuery, linkArgs, _ := psql.
			Insert("recipe_ingredients").
			Columns("recipe_id", "ingredient_id", "amount", "unit").
			Values(recipeID, ingredientIDs[ingredient.Name], 0, ingredient.Measurement).
			ToSql()

		if _, err := tx.ExecContext(ctx, linkQuery, linkArgs...); err != nil {
			return "", fmt.Errorf("linked failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return recipeID, nil
}

func AdvancedPantrySearch(
	db *sql.DB,
	ctx context.Context,
	pantry []string,
	wishlist []string,
	strictness string,
) ([]model.SearchResult, error) {
	// pantry mode - consider entire pantry + wishlist in available pool
	available := append(pantry, wishlist...)

	return executeAdvancedSearch(db, ctx, cleanList(available), cleanList(wishlist), strictness)
}

func AdvancedShoppingListSearch(
	db *sql.DB,
	ctx context.Context,
	shoppingList []string,
	wishlist []string,
	strictness string,
) ([]model.SearchResult, error) {
	// shopping list mode - only consider specific selection + wishlist
	available := append(shoppingList, wishlist...)

	return executeAdvancedSearch(db, ctx, cleanList(available), cleanList(wishlist), strictness)
}

func IncrementTimesMadeGlobally(db *sql.DB, ctx context.Context, recipeID string) error {
	query, args, err := psql.
		Update("recipes").
		Set("times_made_globally", squirrel.Expr("times_made_globally + 1")).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where("id = ?", recipeID).
		ToSql()
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error incrementing recipe times made: %w", err)
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("recipe not found: %s", recipeID)
	}

	return nil
}

func GetRecipeByID(db *sql.DB, ctx context.Context, recipeID string) (*model.Recipe, error) {
	query, args, err := psql.
		Select(
			"id",
			"title",
			"times_made_globally",
			"COALESCE(description, '')",
			"COALESCE(instructions, '')",
			"author_id",
			"prep_time_min",
			"calories",
			"budget_tier",
			"COALESCE(image_url, '')",
			"created_at",
			"updated_at",
		).
		From("recipes").
		Where("id = ?", recipeID).
		ToSql()
	if err != nil {
		return nil, err
	}

	var recipe model.Recipe
	err = db.
		QueryRowContext(ctx, query, args...).
		Scan(
			&recipe.ID,
			&recipe.Title,
			&recipe.TimesMadeGlobally,
			&recipe.Description,
			&recipe.Instructions,
			&recipe.AuthorID,
			&recipe.PrepTimeMinutes,
			&recipe.Calories,
			&recipe.BudgetTier,
			&recipe.ImageURL,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
		)
	if err != nil {
		return nil, fmt.Errorf("error fetching recipe: %w", err)
	}

	return &recipe, nil
}

// remove duplicate entries in order to make match score more accurate
func cleanList(wishlist []string) []string {
	uniqueList := make(map[string]bool)
	var cleanedList []string

	for _, w := range wishlist {
		w = strings.ToLower(strings.TrimSpace(w))

		if !uniqueList[w] && w != "" {
			uniqueList[w] = true
			cleanedList = append(cleanedList, w)
		}
	}

	return cleanedList
}

func executeAdvancedSearch(
	db *sql.DB,
	ctx context.Context,
	available []string,
	wishlist []string,
	strictness string,
) ([]model.SearchResult, error) {
	/*
	 * when building the subquery, it's important to make sure squirrel doesn't flatten the variables before the main query
	 * flattening before main query will result in "$1" appearing in the query multiple times, throwing an error
	 * PlaceholderFormat() ensures subquery uses "?" until full query is built
	 */
	subBuilder := squirrel.
		Select(
			"ri.recipe_id",
			"COUNT(ri.ingredient_id) AS total_count",
			"COUNT(ri.ingredient_id) FILTER (WHERE LOWER(i.name) = ANY(?)) AS match_count",
			"COUNT(ri.ingredient_id) FILTER (WHERE LOWER(i.name) = ANY(?)) AS wishlist_match_count",
		).
		From("recipe_ingredients ri").
		Join("ingredients i ON ri.ingredient_id = i.id").
		GroupBy("ri.recipe_id").
		PlaceholderFormat(squirrel.Question)

	subQuery, _, err := subBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build subquery: %w", err)
	}

	mainQuery := psql.
		Select("r.id", "r.title", "stats.match_count", "stats.total_count").
		From("recipes r").
		Join(fmt.Sprintf("(%s) stats ON r.id = stats.recipe_id", subQuery), pq.Array(available), pq.Array(wishlist)).
		Where("stats.wishlist_match_count = ?", len(wishlist)). // 100% wishlist match must be met for all search thresholds
		OrderBy("(CAST(stats.match_count AS FLOAT) / stats.total_count) DESC")

	var threshold float64
	switch strictness {
	case "strict":
		threshold = model.StrictThreshold
	case "less_strict":
		threshold = model.LessStrictThreshold
	case "unrestricted":
		threshold = model.UnrestrictedTreshold
	}

	mainQuery = mainQuery.Where("CAST(stats.match_count AS FLOAT) / stats.total_count >= ?", threshold)

	query, args, err := mainQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build main query: %w", err)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("[SQL: %s] -> ARGS: %v", query, args)
	}
	defer rows.Close()

	var recipes []model.SearchResult
	for rows.Next() {
		var r model.SearchResult

		if err := rows.Scan(&r.ID, &r.Title, &r.MatchCount, &r.TotalNeeded); err != nil {
			return nil, fmt.Errorf("failed to map query to SearchResult: %w", err)
		}

		recipes = append(recipes, r)
	}

	return recipes, nil
}
