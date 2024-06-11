import { defineStore } from "pinia";

export const useStats = defineStore('stats', {
    state: () => ({
        stats: null
    }),
  
    actions: {
      async fetchStats(from, to) {
        console.log(from)
        try {
            const urlParams = new URLSearchParams();
            urlParams.set('start', from);
            urlParams.set('end', to);
            const resp = await fetch(window.config.backendUrl + '/stats?' + urlParams.toString());
            this.stats = await resp.json();
        } catch (err) {
            console.error(err);
        }
      },
    },
  });