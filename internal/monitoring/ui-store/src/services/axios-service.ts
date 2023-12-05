import axios, {AxiosRequestConfig, AxiosResponse} from 'axios';

class AxiosService {
    private static instance: AxiosService;

    private constructor() {
        // axios.defaults.baseURL = 'http://your-api-url/';
        axios.defaults.headers.post['Content-Type'] = 'application/json';
        axios.defaults.url = '/api/v1/data/';

        const token = localStorage.getItem('token');
        if (token) {
            this.setToken(token);
        }
    }

    public static getInstance(): AxiosService {
        if (!AxiosService.instance) {
            AxiosService.instance = new AxiosService();
        }

        return AxiosService.instance;
    }

    public setToken(token: string): void {
        axios.defaults.headers.post['Authorization'] = `Bearer ${token}`;
    }

    public get<T = any, R = AxiosResponse<T>>(url: string, config?: AxiosRequestConfig): Promise<R> {
        return axios.get<T, R>(url, config);
    }

    public post<T = any, R = AxiosResponse<T>>(url: string, data?: T, config?: AxiosRequestConfig): Promise<R> {
        return axios.post<T, R>(url, data, config);
    }

    public put<T = any, R = AxiosResponse<T>>(url: string, data?: T, config?: AxiosRequestConfig): Promise<R> {
        return axios.put<T, R>(url, data, config);
    }

    public patch<T = any, R = AxiosResponse<T>>(url: string, data?: T, config?: AxiosRequestConfig): Promise<R> {
        return axios.patch<T, R>(url, data, config);
    }

    public delete<T = any, R = AxiosResponse<T>>(url: string, config?: AxiosRequestConfig): Promise<R> {
        return axios.delete<T, R>(url, config);
    }
}

export default AxiosService;