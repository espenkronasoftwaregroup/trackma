import { onMounted, ref } from 'vue'
import { defineStore } from 'pinia'

export const useConfigStore = defineStore('config', () => {
    const config = ref(null);

    async function loadConfig() {
        const response = await fetch('/config.json');
        const cfg = await response.json();
        config.value = cfg; 
        console.log(cfg)
    }

    onMounted(async () => {
        if (config) return;
        await loadConfig();
    });
})