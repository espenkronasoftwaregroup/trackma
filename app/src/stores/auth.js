import { defineStore } from "pinia";

export const useAuth = defineStore('user', {
    state: () => ({
        jwt: null,
        refreshToken: null
    }),

    actions: {
        async login(email, password) {
            this.loading  = true;
            try {
                const resp = await fetch(window.config.backendUrl + '/login', {
                    body: JSON.stringify({
                        email,
                        password
                    })
                });

                if (resp.status === 200) {
                    const data = await resp.json();
                    this.jwt = data.jwt;
                    this.refreshToken = data.refreshToken;
                }
            } catch (err) {
                console.error(err);
            }

            this.loading = false;
        },

        async refresh(refreshToken) {
            this.loading  = true;
            try {
                const resp = await fetch(window.config.backendUrl + '/login', {

                });
            } catch (err) {
                console.error(err);
            }

            this.loading = false;
        },

        async logout() {

        }
    },
});