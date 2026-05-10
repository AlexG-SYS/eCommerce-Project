package data

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Product struct {
	ProductID     int64     `json:"product_id"`
	CategoryID    int64     `json:"category_id"`
	CategoryName  string    `json:"category_name,omitempty"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	IsGstEligible bool      `json:"is_gst_eligible"`
	Variants      []Variant `json:"variants,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type Variant struct {
	VariantID      int64               `json:"variant_id"`
	ProductID      int64               `json:"product_id"`
	SKU            string              `json:"sku"`
	SizeAttr       string              `json:"size_attr"`
	ColorAttr      string              `json:"color_attr"`
	ImageURL       string              `json:"image_url,omitempty"`
	CostPrice      float64             `json:"cost_price,omitempty"`
	SellingPrice   float64             `json:"selling_price"`
	TotalInventory int                 `json:"total_inventory,omitempty"`
	InventoryLocs  []InventoryLocation `json:"inventory_locations,omitempty"`
	CreatedAt      time.Time           `json:"created_at,omitempty"`
}

type InventoryLocation struct {
	LocationName string `json:"location_name"`
	Stock        int    `json:"stock"`
}

type ProductModel struct {
	DB *sql.DB
}

// Validation logic
func ValidateProduct(p *Product) map[string]string {
	errs := make(map[string]string)
	if p.Name == "" {
		errs["name"] = "must be provided"
	}
	if p.CategoryID == 0 {
		errs["category"] = "must be provided"
	}

	return errs
}

func ValidateVariant(v *Variant) map[string]string {
	errs := make(map[string]string)
	if v.SKU == "" {
		errs["sku"] = "must be provided"
	}
	if v.CostPrice < 0 {
		errs["cost_price"] = "must be a non-negative value"
	}
	if v.SellingPrice < 0 {
		errs["selling_price"] = "must be a non-negative value"
	}

	return errs
}

func (m ProductModel) InsertProduct(p *Product) error {
	// ECO PROMPT: Adding a new product to the catalog...
	query := `INSERT INTO products (name, description, category_id, is_gst_eligible) 
			  VALUES ($1, $2, $3, $4) RETURNING product_id, created_at, updated_at`
	return m.DB.QueryRow(query, p.Name, p.Description, p.CategoryID, p.IsGstEligible).Scan(&p.ProductID, &p.CreatedAt, &p.UpdatedAt)
}

func (m ProductModel) InsertVariant(v *Variant) error {
	// ECO PROMPT: Adding a new variant for an existing product...
	query := `INSERT INTO product_variants (product_id, sku, size_attr, color_attr, cost_price, selling_price) 
			  VALUES ($1, $2, $3, $4, $5, $6) RETURNING variant_id, created_at`
	return m.DB.QueryRow(query, v.ProductID, v.SKU, v.SizeAttr, v.ColorAttr, v.CostPrice, v.SellingPrice).Scan(&v.VariantID, &v.CreatedAt)
}

func (m ProductModel) GetProduct(id int64) (*Product, error) {

	var product Product

	// ------------------------------------------
	// 1. Load Product
	// ------------------------------------------

	productQuery := `
		SELECT
			p.product_id,
			p.category_id,
			c.name,
			p.name,
			p.description,
			p.is_gst_eligible,
			p.created_at,
			p.updated_at
		FROM products p
		LEFT JOIN categories c
			ON c.category_id = p.category_id
		WHERE p.product_id = $1
	`

	err := m.DB.QueryRow(productQuery, id).Scan(
		&product.ProductID,
		&product.CategoryID,
		&product.CategoryName,
		&product.Name,
		&product.Description,
		&product.IsGstEligible,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// ------------------------------------------
	// 2. Load Variants
	// ------------------------------------------

	variantQuery := `
		SELECT
			variant_id,
			product_id,
			sku,
			size_attr,
			color_attr,
			image_url,
			cost_price,
			selling_price,
			created_at
		FROM product_variants
		WHERE product_id = $1
		ORDER BY variant_id ASC
	`

	rows, err := m.DB.Query(variantQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []Variant

	for rows.Next() {

		var variant Variant

		err := rows.Scan(
			&variant.VariantID,
			&variant.ProductID,
			&variant.SKU,
			&variant.SizeAttr,
			&variant.ColorAttr,
			&variant.ImageURL,
			&variant.CostPrice,
			&variant.SellingPrice,
			&variant.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		// ------------------------------------------
		// 3. Load Inventory Locations
		// ------------------------------------------

		inventoryQuery := `
			SELECT
				l.name,
				(i.stock_on_hand - i.stock_reserved) AS stock
			FROM inventory i
			JOIN locations l
				ON l.location_id = i.location_id
			WHERE i.variant_id = $1
		`

		invRows, err := m.DB.Query(
			inventoryQuery,
			variant.VariantID,
		)

		if err != nil {
			return nil, err
		}

		totalInventory := 0

		for invRows.Next() {

			var inv InventoryLocation

			err := invRows.Scan(
				&inv.LocationName,
				&inv.Stock,
			)

			if err != nil {
				invRows.Close()
				return nil, err
			}

			totalInventory += inv.Stock

			variant.InventoryLocs =
				append(variant.InventoryLocs, inv)
		}

		invRows.Close()

		variant.TotalInventory = totalInventory

		variants = append(variants, variant)
	}

	product.Variants = variants

	return &product, nil
}

func (m ProductModel) GetVariant(id int64) (*Variant, error) {
	var v Variant
	query := `SELECT variant_id, product_id, sku, size_attr, color_attr, cost_price, selling_price, created_at 
			  FROM product_variants WHERE variant_id = $1`
	err := m.DB.QueryRow(query, id).Scan(
		&v.VariantID, &v.ProductID, &v.SKU, &v.SizeAttr, &v.ColorAttr, &v.CostPrice, &v.SellingPrice, &v.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (m ProductModel) GetVariantByProduct(productID int64) ([]*Variant, error) {
	query := `SELECT variant_id, product_id, sku, size_attr, color_attr, cost_price, selling_price, created_at 
			  FROM product_variants WHERE product_id = $1`
	rows, err := m.DB.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []*Variant
	for rows.Next() {
		var v Variant
		err := rows.Scan(
			&v.VariantID, &v.ProductID, &v.SKU, &v.SizeAttr, &v.ColorAttr, &v.CostPrice, &v.SellingPrice, &v.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		variants = append(variants, &v)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return variants, nil
}

func (m ProductModel) GetAllProducts(name string, categoryID int, f Filters) ([]*Product, Metadata, error) {
	// Construct the query with dynamic sorting
	query := fmt.Sprintf(`
    SELECT count(*) OVER(), 
        p.product_id, 
        p.name, 
        COALESCE(json_agg(json_build_object(
            'variant_id', pv.variant_id,
            'color_attr', pv.color_attr,
			'size_attr', pv.size_attr,
            'selling_price', pv.selling_price,
			'sku', pv.sku,
			'image_url', pv.image_url
        ) ORDER BY pv.variant_id) FILTER (WHERE pv.variant_id IS NOT NULL), '[]') as variants
    FROM products p
    LEFT JOIN product_variants pv ON p.product_id = pv.product_id
    WHERE (to_tsvector('simple', p.name) @@ plainto_tsquery('simple', $1) OR $1 = '')
    AND (p.category_id = $2 OR $2 = 0)
    GROUP BY p.product_id
    ORDER BY %s %s, p.product_id ASC
    LIMIT $3 OFFSET $4`, f.sortColumn(), f.sortDirection())

	args := []any{name, categoryID, f.limit(), f.offset()}

	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	products := []*Product{}

	for rows.Next() {
		var p Product
		var variantsJSON []byte

		err := rows.Scan(
			&totalRecords,
			&p.ProductID,
			&p.Name,
			&variantsJSON,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Unmarshal the JSON bytes into the product's Variants slice
		err = json.Unmarshal(variantsJSON, &p.Variants)
		if err != nil {
			return nil, Metadata{}, err
		}

		products = append(products, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Calculate metadata using the totalRecords
	metadata := CalculateMetadata(totalRecords, f.Page, f.PageSize)

	return products, metadata, nil
}

func (m ProductModel) GetAllVariants(productID int64) ([]*Variant, error) {
	query := `SELECT variant_id, product_id, sku, size_attr, color_attr, cost_price, selling_price, created_at 
			  FROM variants WHERE product_id = $1`
	rows, err := m.DB.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []*Variant
	for rows.Next() {
		var v Variant
		err := rows.Scan(
			&v.VariantID,
			&v.ProductID,
			&v.SKU,
			&v.SizeAttr,
			&v.ColorAttr,
			&v.CostPrice,
			&v.SellingPrice,
			&v.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		variants = append(variants, &v)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return variants, nil
}

// Update uses a pointer (*Product) to modify the record in memory and return new values
func (m ProductModel) UpdateProduct(p *Product) error {
	query := `UPDATE products 
              SET name = $1, category_id = $2, description = $3, is_gst_eligible = $4, updated_at = now()
              WHERE product_id = $5 
              RETURNING updated_at`

	args := []any{p.Name, p.CategoryID, p.Description, p.IsGstEligible, p.ProductID}

	err := m.DB.QueryRow(query, args...).Scan(&p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("edit failed: record not found")
		}
		return err
	}
	return nil
}

func (m ProductModel) UpdateVariant(v *Variant) error {
	query := `
        UPDATE product_variants 
        SET sku = $1, size_attr = $2, color_attr = $3, cost_price = $4, selling_price = $5
        WHERE variant_id = $6
        RETURNING variant_id`

	args := []any{v.SKU, v.SizeAttr, v.ColorAttr, v.CostPrice, v.SellingPrice, v.VariantID}

	return m.DB.QueryRow(query, args...).Scan(&v.VariantID)
}

// Data Should Never be Deleted,
// func (m ProductModel) Delete(id int64) error {
// 	query := `DELETE FROM products WHERE id = $1`
// 	result, err := m.DB.Exec(query, id)
// 	if err != nil {
// 		return err
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}

// 	if rowsAffected == 0 {
// 		return errors.New("record not found")
// 	}

// 	return nil
// }
