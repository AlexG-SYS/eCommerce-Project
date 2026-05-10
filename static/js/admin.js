import { renderAdmin } from './admin-render.js';
import { stateAdmin }  from './admin-state.js';
import { AdminService } from './modules/data-service.js'; 
import { emitter }      from './modules/event-emitter.js';

// Listen for successful login from the DataService
emitter.on('admin:login-success', (data) => {
    stateAdmin.session.isAuthenticated = true;
    stateAdmin.session.user = data.user;   // Contains full_name, role, etc.
    stateAdmin.session.token = data.token; // Store your JWT/Session token
    stateAdmin.loading = false;
    stateAdmin.error = null;
    
    // Once logged in, immediately fetch the inventory
    // AdminService.fetchInventory(); 
    renderAdmin(); 
});

// Listen for login errors (Invalid credentials, server down, etc.)
emitter.on('admin:login-error', (errorMsg) => {
    stateAdmin.loading = false;
    stateAdmin.error = errorMsg;
    renderAdmin();
});

// Initial Load
renderAdmin();

// Attach login handler to window so the HTML 'onsubmit' can find it
window.handleAdminLogin = (e) => {
    e.preventDefault();
    stateAdmin.loading = true;
    stateAdmin.error = null;
    renderAdmin(); // Re-render to show loading state
    const formData = new FormData(e.target);
    const credentials = {
        email: formData.get('username'),
        password: formData.get('password')
    };
    
    // Call the Go Backend via AdminService
    AdminService.login(credentials);
}



window.handleLogout = () => {
    stateAdmin.session.isAuthenticated = false;
    stateAdmin.session.user = null;
    stateAdmin.session.token = null;
    stateAdmin.inventory = [];
    renderAdmin();
};