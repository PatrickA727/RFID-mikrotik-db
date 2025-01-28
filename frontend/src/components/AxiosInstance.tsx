import axios, { AxiosError } from "axios"

const api = axios.create({
    baseURL: 'http://localhost:5000',        
    withCredentials: true, 
    headers: {
        'Content-Type': 'application/json',
      },
});

api.interceptors.response.use(  
    (response) => response, // Successful API's are ignored

    async (error: AxiosError) => {
        const originalRequest = error.config;   // Gets the data/config for the request that failed/error

        if (originalRequest && error.response?.status === 403 && !originalRequest._retry) {
            originalRequest._retry = true;

            try {
                await axios.post('/api/user/refresh', { withCredentials: true });
        
                return api(originalRequest);  // Retry the original request
            } catch (refreshError) {
                console.error('Refresh token failed, logging out...');
                // window.location.href = '/';     // Navigate to login page
                return Promise.reject(refreshError);    // Reject request on refresh fail
              }
        }

        return Promise.reject(error); 
    }
)

export default api