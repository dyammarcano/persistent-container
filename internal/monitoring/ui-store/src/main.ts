import {createApp} from 'vue'
import {createPinia} from 'pinia'
import './styles.scss'

import 'module-alias/register';

import App from './App.vue'
import router from './router'

createApp(App)
    .use(createPinia())
    .use(router)
    .mount('#app')
