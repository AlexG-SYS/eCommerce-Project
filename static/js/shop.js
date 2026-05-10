import { emitter }     from './modules/event-emitter.js';
import { DataService } from './modules/data-service.js';
import { render }      from './shop-render.js';
import { state }       from './shop-state.js';

// 1. Listen for successful data load
emitter.on('products:loaded', (data) => {
  state.products = data.products;
  state.metadata = data.metadata;
  state.loading = false;
  state.paginationLoading = false;
  state.error = null;
  render(); 
});

// 2. Listen for errors
emitter.on('products:error', (errorMsg) => {
  state.loading = false;
  state.paginationLoading = false;
  state.error = errorMsg;
  render();
});

// 3. Global helper for the UI buttons
window.openProductInfo = async (productId) => {

    try {
        state.productModalLoading = true;
        state.selectedProduct = null;
        render();

        // NEW API CALL
        const product =
            await DataService.fetchProduct(productId);
        state.selectedProduct = product;

    } catch (err) {
        console.error(err);
        alert("Unable to load product");

    } finally {
        state.productModalLoading = false;
        render();
    }
};

// Start the app
state.loading = true; // Set loading before fetch
render(); 
DataService.fetchProducts();

window.updateVariantSelection = function(productId, variantId) {
    const product = state.products.find(p => p.product_id === productId);
    const variant = product.variants.find(v => v.variant_id === variantId);
    
    const container = document.querySelector(`#product-${productId}`);
    if (!container || !variant) return;

    // Update Text
    container.querySelector('.variant-name-display').innerText = `${variant.color_attr} Edition`;
    container.querySelector('.variant-size-display').innerText = `${variant.size_attr}`;
    container.querySelector('.variant-price-display').innerText = `$${variant.selling_price.toFixed(2)}`;

    // Update Active State on Circles
    container.querySelectorAll('.color-circle').forEach(btn => btn.classList.remove('active'));
    event.target.classList.add('active');
};

// Global function to handle page clicks
window.changePage = (newPage) => {
    // 1. Set loading state so the UI shows the spinner
    if (state.paginationLoading) return;

    state.paginationLoading = true;
    render();
    
    // 2. Trigger the fetch with the specific page number
    DataService.fetchProducts(newPage);
    
    // 3. Optional: Smooth scroll back to top so user sees the new items
    window.scrollTo({ top: 0, behavior: 'smooth' });
};

// 1. Increment Cart
window.dispatchAddToCart = (productId) => {
  // Increase count in state
  state.cartCount++;
  
  console.log(`[Cart] Total items: ${state.cartCount}`);
  
  // Re-render to update the header number
  render();
};



// 3. Close Product Info
window.closeProductInfo = () => {
    state.selectedProduct = null;
    render();
};