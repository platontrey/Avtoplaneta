package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"

	"parts-service/internal/catalog"
	"parts-service/internal/domain"
	"parts-service/internal/search"
)

// ─── FindWithFilters (Squirrel dynamic query) ───────────────────────────────

func buildCategoryPredicates(value interface{}) squirrel.Sqlizer {
	if variants, ok := value.([]string); ok {
		conds := make([]squirrel.Sqlizer, 0, len(variants))
		for _, v := range variants {
			if strings.TrimSpace(v) != "" {
				conds = append(conds, squirrel.Expr("category ILIKE ?", "%"+v+"%"))
			}
		}
		if len(conds) > 0 {
			return squirrel.Or(conds)
		}
		return nil
	}
	if str, ok := value.(string); ok && strings.TrimSpace(str) != "" {
		return squirrel.Expr("category ILIKE ?", "%"+str+"%")
	}
	return nil
}

func buildSynonymPredicates(col string, synonyms []string) squirrel.Sqlizer {
	conds := make([]squirrel.Sqlizer, 0, len(synonyms))
	for _, syn := range synonyms {
		if len(syn) == 1 {
			conds = append(conds, squirrel.Eq{col: syn})
		} else {
			conds = append(conds, squirrel.Expr(col+" ILIKE ?", "%"+syn+"%"))
		}
	}
	return squirrel.Or(conds)
}

func buildCodePredicate(col string, value interface{}) squirrel.Sqlizer {
	val := strings.TrimSpace(value.(string))
	cleanVal := search.CleanPartNumber(val)
	if cleanVal != "" {
		return squirrel.Or{
			squirrel.Eq{col: val},
			squirrel.Expr(fmt.Sprintf("REPLACE(REPLACE(LOWER(%s), '-', ''), ' ', '') = ?", col), cleanVal),
		}
	}
	return squirrel.Eq{col: val}
}

func buildSearchTermSqlizer(term string) squirrel.Sqlizer {
	termPattern := "%" + term + "%"
	baseExpr := "(name ILIKE ? OR description ILIKE ? OR brand ILIKE ? OR model ILIKE ? OR number ILIKE ? OR oem_code ILIKE ? OR vin ILIKE ? OR category ILIKE ? OR car_release_date ILIKE ? OR car_release_period ILIKE ? OR body_brand ILIKE ? OR engine_brand ILIKE ? OR front_rear ILIKE ? OR left_right ILIKE ? OR top_bottom ILIKE ? OR color ILIKE ? OR condition ILIKE ? OR transmission ILIKE ? OR transmission_model ILIKE ? OR drive ILIKE ? OR defect ILIKE ? OR wear_percentage ILIKE ? OR season ILIKE ? OR diameter ILIKE ? OR width ILIKE ? OR profile ILIKE ? OR drilling ILIKE ? OR \"offset\" ILIKE ? OR center_hole_diameter ILIKE ? OR tire_model ILIKE ? OR tire_quantity ILIKE ? OR location ILIKE ? OR address ILIKE ? OR salesman ILIKE ? OR manufacturer ILIKE ? OR manufacturer_code ILIKE ? OR supplier_code ILIKE ?"
	args := make([]interface{}, 37)
	for i := range args {
		args[i] = termPattern
	}

	cleanTerm := strings.ToLower(term)
	if strings.HasPrefix(cleanTerm, "перед") || cleanTerm == "front" {
		baseExpr += " OR front_rear = 'F' OR front_rear ILIKE '%перед%'"
	} else if strings.HasPrefix(cleanTerm, "зад") || cleanTerm == "rear" {
		baseExpr += " OR front_rear = 'R' OR front_rear ILIKE '%зад%'"
	}
	if strings.HasPrefix(cleanTerm, "прав") || cleanTerm == "right" {
		baseExpr += " OR left_right = 'R' OR left_right ILIKE '%прав%'"
	} else if strings.HasPrefix(cleanTerm, "лев") || cleanTerm == "left" {
		baseExpr += " OR left_right = 'L' OR left_right ILIKE '%лев%'"
	}
	if strings.HasPrefix(cleanTerm, "верх") || cleanTerm == "top" || cleanTerm == "upper" {
		baseExpr += " OR top_bottom = 'T' OR top_bottom = 'U' OR top_bottom ILIKE '%верх%'"
	} else if strings.HasPrefix(cleanTerm, "низ") || cleanTerm == "bottom" || cleanTerm == "lower" {
		baseExpr += " OR top_bottom = 'B' OR top_bottom = 'L' OR top_bottom ILIKE '%низ%'"
	}
	baseExpr += ")"
	return squirrel.Expr(baseExpr, args...)
}

func buildFragmentAndPred(frag string) squirrel.Sqlizer {
	var andPreds []squirrel.Sqlizer
	for _, term := range strings.Fields(frag) {
		if strings.TrimSpace(term) != "" {
			andPreds = append(andPreds, buildSearchTermSqlizer(term))
		}
	}
	if len(andPreds) == 0 {
		return nil
	}
	return squirrel.And(andPreds)
}

func buildSearchPredicates(searchStr string) []squirrel.Sqlizer {
	searchStr = strings.TrimSpace(searchStr)
	if searchStr == "" {
		return nil
	}

	fragments := search.SplitMultiSearchQuery(searchStr)
	if len(fragments) > 1 {
		var orPreds []squirrel.Sqlizer
		for _, frag := range fragments {
			if pred := buildFragmentAndPred(frag); pred != nil {
				orPreds = append(orPreds, pred)
			}
		}
		if len(orPreds) > 0 {
			return []squirrel.Sqlizer{squirrel.Or(orPreds)}
		}
		return nil
	}

	if len(fragments) == 1 {
		var preds []squirrel.Sqlizer
		for _, term := range strings.Fields(fragments[0]) {
			if strings.TrimSpace(term) != "" {
				preds = append(preds, buildSearchTermSqlizer(term))
			}
		}
		return preds
	}

	return nil
}

func (r *partRepository) buildFilterPredicates(filters map[string]interface{}) []squirrel.Sqlizer {
	var preds []squirrel.Sqlizer

	for key, value := range filters {
		switch key {
		case "ids_in":
			preds = append(preds, squirrel.Eq{"id": value})
		case "to_delete_at_is_null":
			if value.(bool) {
				preds = append(preds, squirrel.Expr("to_delete_at IS NULL"))
			}
		case "quantity_gte":
			preds = append(preds, squirrel.GtOrEq{"quantity": value})
		case "category_ilike":
			if pred := buildCategoryPredicates(value); pred != nil {
				preds = append(preds, pred)
			}
		case "brand_ilike":
			preds = append(preds, squirrel.Expr("brand ILIKE ?", "%"+value.(string)+"%"))
		case "model_ilike":
			preds = append(preds, squirrel.Expr("model ILIKE ?", "%"+value.(string)+"%"))
		case "location_ilike":
			preds = append(preds, squirrel.Expr("location ILIKE ?", "%"+value.(string)+"%"))
		case "address_ilike":
			preds = append(preds, squirrel.Expr("address ILIKE ?", "%"+value.(string)+"%"))
		case "salesman_ilike":
			preds = append(preds, squirrel.Expr("salesman ILIKE ?", "%"+value.(string)+"%"))
		case "status":
			preds = append(preds, squirrel.Eq{"status": value})
		case "has_photo":
			if value.(bool) {
				preds = append(preds, squirrel.Expr("photos IS NOT NULL AND CASE WHEN jsonb_typeof(photos) = 'array' THEN jsonb_array_length(photos) > 0 ELSE false END"))
			} else {
				preds = append(preds, squirrel.Expr("(photos IS NULL OR CASE WHEN jsonb_typeof(photos) = 'array' THEN jsonb_array_length(photos) = 0 ELSE true END)"))
			}
		case "number_ilike":
			preds = append(preds, buildCodePredicate("number", value))
		case "oem_code_ilike":
			preds = append(preds, buildCodePredicate("oem_code", value))
		case "vin_ilike":
			preds = append(preds, squirrel.Expr("vin ILIKE ?", "%"+value.(string)+"%"))
		case "body_brand_ilike":
			preds = append(preds, squirrel.Expr("body_brand ILIKE ?", "%"+value.(string)+"%"))
		case "engine_brand_ilike":
			preds = append(preds, squirrel.Expr("engine_brand ILIKE ?", "%"+value.(string)+"%"))
		case "car_release_date_ilike":
			preds = append(preds, squirrel.Expr("car_release_date ILIKE ?", "%"+value.(string)+"%"))
		case "car_release_period_ilike":
			preds = append(preds, squirrel.Expr("car_release_period ILIKE ?", "%"+value.(string)+"%"))
		case "transmission_ilike":
			preds = append(preds, squirrel.Expr("transmission ILIKE ?", "%"+value.(string)+"%"))
		case "drive_ilike":
			preds = append(preds, squirrel.Expr("drive ILIKE ?", "%"+value.(string)+"%"))
		case "condition_ilike":
			preds = append(preds, squirrel.Expr("condition ILIKE ?", "%"+value.(string)+"%"))
		case "manufacturer_ilike":
			preds = append(preds, squirrel.Expr("manufacturer ILIKE ?", "%"+value.(string)+"%"))
		case "defect_ilike":
			preds = append(preds, squirrel.Expr("defect ILIKE ?", "%"+value.(string)+"%"))
		case "color_ilike":
			preds = append(preds, squirrel.Expr("color ILIKE ?", "%"+value.(string)+"%"))
		case "min_price":
			preds = append(preds, squirrel.GtOrEq{"price": value})
		case "max_price":
			preds = append(preds, squirrel.LtOrEq{"price": value})
		case "min_quantity":
			preds = append(preds, squirrel.GtOrEq{"quantity": value})
		case "max_quantity":
			preds = append(preds, squirrel.LtOrEq{"quantity": value})
		case "front_rear_ilike":
			preds = append(preds, buildSynonymPredicates("front_rear", catalog.ExpandFrontRearSynonyms(value.(string))))
		case "left_right_ilike":
			preds = append(preds, buildSynonymPredicates("left_right", catalog.ExpandLeftRightSynonyms(value.(string))))
		case "top_bottom_ilike":
			preds = append(preds, buildSynonymPredicates("top_bottom", catalog.ExpandTopBottomSynonyms(value.(string))))
		case "manufacturer_code_ilike":
			preds = append(preds, buildCodePredicate("manufacturer_code", value))
		case "supplier_code_ilike":
			preds = append(preds, squirrel.Expr("supplier_code ILIKE ?", "%"+value.(string)+"%"))
		case "transmission_model_ilike":
			preds = append(preds, squirrel.Expr("transmission_model ILIKE ?", "%"+value.(string)+"%"))
		case "wear_percentage_ilike":
			preds = append(preds, squirrel.Expr("wear_percentage ILIKE ?", "%"+value.(string)+"%"))
		case "season_ilike":
			preds = append(preds, squirrel.Expr("season ILIKE ?", "%"+value.(string)+"%"))
		case "diameter_ilike":
			preds = append(preds, squirrel.Expr("diameter ILIKE ?", "%"+value.(string)+"%"))
		case "width_ilike":
			preds = append(preds, squirrel.Expr("width ILIKE ?", "%"+value.(string)+"%"))
		case "profile_ilike":
			preds = append(preds, squirrel.Expr("profile ILIKE ?", "%"+value.(string)+"%"))
		case "tire_quantity_ilike":
			preds = append(preds, squirrel.Expr("tire_quantity ILIKE ?", "%"+value.(string)+"%"))
		case "drilling_ilike":
			preds = append(preds, squirrel.Expr("drilling ILIKE ?", "%"+value.(string)+"%"))
		case "offset_ilike":
			preds = append(preds, squirrel.Expr("\"offset\" ILIKE ?", "%"+value.(string)+"%"))
		case "center_hole_diameter_ilike":
			preds = append(preds, squirrel.Expr("center_hole_diameter ILIKE ?", "%"+value.(string)+"%"))
		case "tire_model_ilike":
			preds = append(preds, squirrel.Expr("tire_model ILIKE ?", "%"+value.(string)+"%"))
		case "search":
			if sp := buildSearchPredicates(value.(string)); len(sp) > 0 {
				preds = append(preds, sp...)
			}
		}
	}
	return preds
}

func (r *partRepository) FindWithFilters(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]domain.Part, error) {
	builder := r.psq.Select(partColumns).From("parts").Where("deleted_at IS NULL")

	for _, pred := range r.buildFilterPredicates(filters) {
		builder = builder.Where(pred)
	}

	builder = builder.OrderBy("id DESC")

	if limit > 0 {
		builder = builder.Limit(uint64(limit)).Offset(uint64(offset))
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanParts(rows)
}

func (r *partRepository) CountWithFilters(ctx context.Context, filters map[string]interface{}) (int64, error) {
	builder := r.psq.Select("COUNT(*)").From("parts").Where("deleted_at IS NULL")

	for _, pred := range r.buildFilterPredicates(filters) {
		builder = builder.Where(pred)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var count int64
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
