<script setup>
    const props = defineProps(['value', 'label']);
    const formatter = new Intl.NumberFormat('sv-SE');
</script>

<template>
    <div class="has-text-centered">
        <div>
            <p class="heading">{{ label }}</p>
            <p class="title">{{ formatter.format(value) }}</p>
        </div>
    </div>
</template>