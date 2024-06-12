<script setup>
import { ref } from 'vue';

const props = defineProps(['label']);
const model = defineModel();

</script>

<template>
    <div class="field is-horizontal">
        <div class="field-label is-normal">
            <label class="label">{{label}}</label>
        </div>
        <div class="field-body">
            <div class="field">
                <p class="control">
                    <input class="input" type="date" v-model="model" />
                </p>
            </div>
        </div>
    </div>
</template>