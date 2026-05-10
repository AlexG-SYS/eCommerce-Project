export const stateAdmin = {
  inventory: [],
  metadata: {},
  loading: false,
  error: null,
  filters: {
        page: 1,
        page_size: 10,
        search: ""
},
  session: {
    isAuthenticated: false,
    user: null,
    token: null
  }
};