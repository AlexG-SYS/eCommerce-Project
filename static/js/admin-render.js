import { stateAdmin } from './admin-state.js';

export function renderAdmin() {
    const app = document.querySelector("#app");
    if (!app) return;

    // 1. Check Authentication Status
    if (!stateAdmin.session.isAuthenticated) {
        app.innerHTML = `
            <div class="admin-login-container">
                <div class="login-card">
                    <div class="login-header">
                        <h2>Ace Closet Admin</h2>
                        <p>Please sign in to manage your store</p>
                    </div>

                    <form class="ui-form" onsubmit="window.handleAdminLogin(event)">
                        <div class="input-group">
                            <label class="header-label">Username</label>
                            <input type="text" name="username" class="ui-input" required placeholder="Enter username">
                        </div>

                        <div class="input-group" style="margin-top: 1rem; margin-bottom: 1rem;">
                            <label class="header-label">Password</label>
                            <input type="password" name="password" class="ui-input" required placeholder="••••••••">
                        </div>

                        ${stateAdmin.error ? `<div class="status-error">${stateAdmin.error}</div>` : ''}

                        <button type="submit" class="product-btn" style="margin-top: 1rem;" ${stateAdmin.loading ? 'disabled' : ''}>
                            ${stateAdmin.loading ? '<span class="spinner"></span> Verifying...' : 'Login'}
                        </button>
                    </form>
                </div>
            </div>
        `;
        return; 
    }

    // Inside your renderAdmin function, where the dashboard is defined:
const userName = stateAdmin.session.user ? stateAdmin.session.user.full_name : "Admin";

    // 2. If Authenticated, show the Dashboard Shell
    app.innerHTML = `
    <header class="admin-header">
        <div class="header-inner">
            <div class="logo-section">
                <h1>Ace Closet <span class="badge">Admin</span></h1>
            </div>
            <div class="user-info">
                <span>Logged in as: <strong>${userName}</strong></span>
                <button onclick="handleLogout()" class="page-btn">Logout</button>
            </div>
        </div>
    </header>
    <main>
        </main>
`;
}