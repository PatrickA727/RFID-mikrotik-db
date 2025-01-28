// axios.d.ts
import { AxiosRequestConfig } from 'axios';

// Extend the AxiosRequestConfig interface
declare module 'axios' {
    interface AxiosRequestConfig {
        _retry?: boolean;  
    }
}
