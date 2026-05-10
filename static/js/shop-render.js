import { state } from "./shop-state.js";

function renderProduct(product) {
  const variants = product.variants || [];
  const firstVariant = variants[0] || {};

  return `
    <div class="product-card" id="product-${product.product_id}">
      
      <div class="product-image" onclick="window.openProductInfo(${product.product_id})" style="cursor:pointer">
        <img src="${product.variants?.[0]?.image_url }" alt="${product.name}">
      </div>  
      <div class="product-info" >
        
        <h3 onclick="window.openProductInfo(${product.product_id})" style="cursor:pointer">${product.name}</h3>

         <div class="variant-details" id="details-${product.product_id}">
            <p class="variant-name-display">${firstVariant.color_attr || ""} Edition</p>
            <p class="variant-size-display">${firstVariant.size_attr || "N/A"}</p>
            <p class="variant-price-display">$${Number(firstVariant.selling_price).toFixed(2) || "0.00"}</p>
          </div>
        
        <div class="ui-form">
          <div class="color-swatch-group">
            <label class="header-label">Select:</label>

            ${variants
      .map(
        (v, index) => `
              <button 
                class="color-circle ${index === 0 ? "active" : ""}" 
                style="background-color: ${v.color_attr.toLowerCase()};"
                onclick="window.updateVariantSelection(${product.product_id}, ${v.variant_id})"
                title="${v.color_attr}"
              ></button>
            `,
      )
      .join("")}
          </div>

          <button 
            class="product-btn" 
            id="add-btn-${product.product_id}"
            onclick="window.dispatchAddToCart(${product.product_id})"
          >
            Add to Cart
          </button>
        </div>
      </div>
    </div>`;
}

function renderPagination(metadata) {
  if (!metadata || metadata.total_records === 0) return "";

  // Calculate the item range (e.g., 1 to 10)
  const start = (metadata.current_page - 1) * metadata.page_size + 1;
  const end = Math.min(
    metadata.current_page * metadata.page_size,
    metadata.total_records,
  );

  return `
    <div class="pagination-section">
      <div class="pagination-info">
        Showing <strong>${start}–${end}</strong> of <strong>${metadata.total_records}</strong> items
      </div>

      <div class="pagination-controls">
        <button 
          class="page-btn" 
          ${metadata.current_page <= metadata.first_page ? "disabled" : ""} 
          onclick="window.changePage(${metadata.current_page - 1})"
        >
          &larr; Previous
        </button>

        <span class="page-indicator">
          Page ${metadata.current_page} of ${metadata.last_page}
        </span>

        <button 
          class="page-btn" 
          ${metadata.current_page >= metadata.last_page ? "disabled" : ""} 
          onclick="window.changePage(${metadata.current_page + 1})"
        >
          Next &rarr;
        </button>
      </div>
    </div>
    `;
}

function renderProductModal(product) {

  const firstImage =
    product.variants?.[0]?.image_url;

  return `
    <div class="modal-overlay"
         onclick="window.closeProductInfo()">

      <div class="modal-content product-modal"
           onclick="event.stopPropagation()">

        <button class="close-btn"
                onclick="window.closeProductInfo()">
          ×
        </button>

        <div class="modal-body">

          <!-- LEFT -->
          <div class="modal-image-section">

            <img
              class="modal-main-image"
              src="${firstImage}"
              alt="${product.name}"
            />

          </div>

          <!-- RIGHT -->
          <div class="modal-details-section">

            <div class="modal-header">

              <span class="modal-category">
                ${product.category_name || "Product"}
              </span>

              <h2>${product.name}</h2>

              <p class="modal-description">
                ${product.description || "No description available."}
              </p>

              <div class="tax-badge">
                ${
                  product.is_gst_eligible
                    ? "GST Applicable"
                    : "GST Exempt"
                }
              </div>

            </div>

            <div class="modal-divider"></div>

            <div class="modal-variants">

              <h4>Available Variants</h4>

              ${product.variants.map(v => `

                <div class="modal-variant-card">

                  <div class="variant-top-row">

                    <div class="variant-color-group">

                      <span
                        class="variant-color-preview"
                        style="background:${v.color_attr?.toLowerCase()}"
                      ></span>

                      <div>
                        <h5>
                          ${v.color_attr || "Default"}
                        </h5>

                        <p>
                          Size:
                          ${v.size_attr || "N/A"}
                        </p>
                      </div>

                    </div>

                    <div class="variant-price">
                      $${Number(v.selling_price).toFixed(2)}
                    </div>

                  </div>

                  <div class="variant-meta">

                    <span>
                      SKU:
                      ${v.sku}
                    </span>

                    <span>
                      ${v.total_inventory || 0} In Stock
                    </span>

                  </div>

                  ${
                    v.inventory_locations?.length
                      ? `
                       
                      `
                      : `
                        <div class="out-stock">
                          Out of stock
                        </div>
                      `
                  }

                  <button
                    class="product-btn modal-cart-btn"
                    onclick="window.dispatchAddToCart(${product.product_id})"
                  >
                    Add to Cart
                  </button>

                </div>

              `).join("")}

            </div>

          </div>

        </div>

      </div>
    </div>
  `;
}

export function render() {
  const app = document.querySelector("#app");
  if (!app) return;

  // 1. Determine what goes inside the <main> section
  let mainContent = "";

  if (state.loading && state.products.length === 0) {
    mainContent = `<div class="status-loading"><span class="spinner"></span>Loading Products...</div>`;
  } else if (state.error) {
    mainContent = `<div class="status-error">${state.error}</div>`;
  } else if (state.products.length === 0) {
    mainContent = `<div class="status-loading">No products found.</div>`;
  } else {
     mainContent = `
    <div class="product-grid ${state.paginationLoading ? 'faded' : ''}">
      ${state.products.map((p) => renderProduct(p)).join("")}
    </div>

    ${state.paginationLoading ? `
      <div class="grid-loading-overlay">
        <div class="spinner"></div>
        <p>Loading page...</p>
      </div>
    ` : renderPagination(state.metadata)}
  `;
  }

  // 2. Render the full page structure with the dynamic mainContent
  app.innerHTML = `
    <header class="main-header">
      <div class="header-inner">
        <div class="logo-section">
          <h1 onclick="window.location.reload()" style="cursor:pointer">Ace Closet</h1>
          <span class="header-label">Belmopan, BZ</span>
        </div>

        <div class="search-section">
          <div class="search-bar">
            <input type="text" placeholder="Search products..." id="product-search">
            <button class="search-btn">🔍</button>
          </div>
        </div>

        <div class="user-section">
          <a href="/profile" class="nav-icon-link">
             <span class="nav-icon">👤</span> Profile
          </a>
         <a href="#" class="nav-icon-link">
          <span class="nav-icon">🛒</span> Cart (${state.cartCount})
        </a>
        </div>
      </div>

      <nav class="category-nav">
        <div class="category-inner">
          <a href="#" class="category-link active">All Products</a>
          <a href="#" class="category-link">Apparel</a>
          <a href="#" class="category-link">Accessories</a>
          <a href="#" class="category-link">Footwear</a>
          <a href="#" class="category-link">New Arrivals</a>
        </div>
      </nav>
    </header>

   <main id="main-content">
      ${mainContent}
    </main>

    <footer class="main-footer">
      <div class="footer-grid">
        <div class="footer-col">
          <h4>Shop</h4>
          <ul>
            <li><a href="#">New Arrivals</a></li>
            <li><a href="#">Best Sellers</a></li>
            <li><a href="#">Clearance</a></li>
          </ul>
        </div>
        <div class="footer-col">
          <h4>Support</h4>
          <ul>
            <li><a href="#">Shipping Policy</a></li>
            <li><a href="#">Returns & Exchanges</a></li>
            <li><a href="#">FAQs</a></li>
          </ul>
        </div>
        <div class="footer-col">
          <h4>Company</h4>
          <ul>
            <li><a href="#">About Us</a></li>
            <li><a href="#">Contact</a></li>
            <li><a href="#">Privacy Policy</a></li>
          </ul>
        </div>
      </div>
      <div class="footer-bottom">
        <p>&copy; 2026 Ace Closet Belize. All rights reserved.</p>
      </div>
    </footer>
    ${
  state.productModalLoading
    ? `
      <div class="modal-overlay">
        <div class="modal-content product-modal loading-modal">

          <div class="modal-body">

            <!-- LEFT SKELETON -->
            <div class="modal-image-section">
              <div class="skeleton-image"></div>
            </div>

            <!-- RIGHT SKELETON -->
            <div class="modal-details-section">

              <div class="skeleton-line w-60"></div>
              <div class="skeleton-line w-40"></div>
              <div class="skeleton-line w-80"></div>

              <div class="skeleton-divider"></div>

              <div class="skeleton-block"></div>
              <div class="skeleton-block"></div>

            </div>

          </div>

        </div>
      </div>
    `
    : state.selectedProduct
      ? renderProductModal(state.selectedProduct)
      : ""
}
  `;
}
