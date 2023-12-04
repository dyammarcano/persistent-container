import {defineStore} from "pinia";
import axios from "axios";

export const useMetricsStore = defineStore({
    id: "metrics",
    state: () => ({
        metrics: [],
    }),
    actions: {
        async fetchMetrics() {
            const response = await axios.get("/api/v1/metrics");
            this.metrics = response.data;
        },
    },
});
