import { emitter } from './event-emitter.js';

const BASE = 'http://localhost:4000/v1';

export const DataService = {
  async fetchProducts(page = 1) {
    try {
      // Create a 1-second delay promise
      const delay = new Promise(resolve => setTimeout(resolve, 1000));

      // 2. Start the fetch request
      const url = `${BASE}/products?page=${page}&page_size=10`;
      const res = await fetch(url);
      if (!res.ok) throw new Error("Failed to fetch products");
      
      // 3. WAIT for the 1-second delay to finish even if the data came back faster
      await delay;

      const data = await res.json();
      // Emit the products so the UI can update
      emitter.emit('products:loaded', data);
    } catch (err) {
      console.error("API Error:", err);
      emitter.emit('products:error', 'Unable to load products. Please try again later.');
    }
  },

  async fetchProduct(productId) {

    try {

      // Create a 1-second delay promise
      const delay = new Promise(resolve => setTimeout(resolve, 1000));

      const res = await fetch(
        `${BASE}/products/${productId}`
      );

      if (!res.ok) {
        throw new Error("Failed to fetch product");
      }

      await delay;

      const data = await res.json();

      return data.product;

    } catch (err) {
      console.error("API Error:", err);
      emitter.emit('products:error', 'Unable to load products. Please try again later.');
    }
  }
};


export const AdminService = {
  async login(credentials) {
    try {
      
      // Professional 1-second delay for UI stability
      const delay = new Promise(resolve => setTimeout(resolve, 1000));
      
      const res = await fetch('http://localhost:4000/v1/admin/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(credentials)
      });
      
      if (!res.ok) {
          // If Go returns 401/403, throw error to catch block
          const errorData = await res.json();
          throw new Error(errorData.error || "Unauthorized");
      }
      
      const data = await res.json();
      await delay;

      // Emit success with the user data 
      emitter.emit('admin:login-success', data);
    } catch (err) {
      emitter.emit('admin:login-error', err.message);
    }
  }
};