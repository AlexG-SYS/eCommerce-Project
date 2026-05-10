export const state = {
    products: [],
    metadata: {}, 
    loading: false,
    error: null,
    cartCount: 0,       
    selectedProduct: null,
    paginationLoading: false,
    productModalLoading: false,
    filters: {
        page: 1,
        page_size: 10,
        search: ""
    }
};