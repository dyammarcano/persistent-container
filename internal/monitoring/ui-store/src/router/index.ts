import {createRouter, createWebHistory, RouteRecordRaw} from "vue-router";
import HomeView from "../views/HomeView.vue";
import ListDataView from "../views/ListDataView.vue";
import CounterView from "../views/CounterView.vue";

export const routes: RouteRecordRaw[] = [
    {
        path: '/',
        name: 'Home',
        component: HomeView
    },
    {
        path: '/list',
        name: 'List',
        component: ListDataView
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