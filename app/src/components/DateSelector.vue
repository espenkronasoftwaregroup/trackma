<script setup>
import { ref } from 'vue';

const props = defineProps(['modelValue', 'label']);
const emit = defineEmits(['update:date']);

</script>

<template>
    <div class="field is-horizontal">
        <div class="field-label is-normal">
            <label class="label">{{label}}</label>
        </div>
        <div class="field-body">
            <div class="field">
                <p class="control">
                    <input class="input" type="date" :value="props.date" @input="emit('update:date', $event.target.value)" />
                </p>
            </div>
        </div>
    </div>
</template>