import {createRouter, createWebHistory, RouteRecordRaw} from "vue-router";
import HomeView from "@/components/HomeView.vue";
import ListDataView from "@/components/ListDataView.vue";
import CounterView from "@/components/CounterView.vue";
import PutDataView from "@/components/PutDataView.vue";

export const routes: RouteRecordRaw[] = [
    {
        path: '/',
        name: 'Home',
        component: HomeView
    },
    {
        path: '/list',
        name: 'List Data',
        component: ListDataView
    },
    {
        path: '/add',
        name: 'Add Data',
        component: PutDataView
    },
    {
        path: '/counter',
        name: 'Counter',
        component: CounterView
    }
]

export const createAppRouter = () => createRouter({
    history: createWebHistory(),
    routes
});

const router = createAppRouter();

export default router;