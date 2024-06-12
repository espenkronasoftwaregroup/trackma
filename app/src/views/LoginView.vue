<script setup>
import { ref } from 'vue';
import { useAuth } from '@/stores/auth.js';

const errorMessage = ref('');
const email = ref('');
const password = ref('');
const auth = useAuth();

async function login() {

    await auth.login(email.value, password.value);

    if (!auth.jwt) {
        errorMessage.value = 'Invalid username or password';
        setTimeout(() => {
            errorMessage.value = '';
        }, 7000);
    }
}

</script>

<template>
    <div class="section">
        <div class="columns is-centered">
            <div class="column is-half box" style="padding: 25px; margin-top: 50px;">
                <div class="field">
                    <label class="label">Email</label>
                    <div class="control">
                        <input type="email" placeholder="bobsmith@gmail.com" class="input" required v-model="email" />
                    </div>
                </div>
                <div class="field">
                    <label class="label">Password</label>
                    <div class="control">
                        <input type="password" placeholder="abc123" class="input" required v-model="password" />
                    </div>
                </div>
                <div class="field">
                    <button @click="login" class="button is-link">
                        Login
                    </button>
                </div>

                <p class="has-text-danger" v-if="errorMessage">{{ errorMessage }}</p>
            </div>
        </div>
    </div>
</template>