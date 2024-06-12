import './assets/bulma.min.css'
import './assets/main.css'

import VueApexCharts from 'vue3-apexcharts'
import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'


fetch ('/config.json')
.then(response => response.json())
.then(cfg => {

    window.config = cfg;
    const app = createApp(App);

    app.use(VueApexCharts);
    app.use(createPinia());
    app.use(router)

    app.mount('#app');
});