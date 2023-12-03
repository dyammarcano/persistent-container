import { defineComponent, ref } from 'vue';
import { Router } from 'vue-router';

export default defineComponent({
    setup(props) {
        const router = props.router;

        return {
            router,
        };
    },
    methods: {
        navigateToHome() {
            router.push('/');
        },
    },
});

// import { createRouter, createWebHistory } from
//
//         'vue-router';
//
// const routes = [
//     {
//         path: '/',
//         name: 'Home',
//         component: () =>
//
//             import('./views/Home.vue'),
//     },
//     {
//         path: '/about',
//         name: 'About',
//         component:
//
//             () =>
//
//                 import('./views/About.vue'),
//     },
// ];
//
// const router = createRouter({
//     history: createWebHistory(),
//     routes,
// });
//
// export
//
// default router;
