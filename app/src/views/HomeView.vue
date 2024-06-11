<script setup>
import DateSelector from '@/components/DateSelector.vue';
import { useStats } from '@/stores/stats.js';
import { ref } from 'vue';

const statsStore = useStats();

const startDate = defineModel(new Date().toISOString().substring(0,10));
const endDate = ref(new Date().toISOString().substring(0, 10));

async function loadData() {
  statsStore.fetchStats(startDate.value, endDate.value);
}

</script>

<template>
  <div class="section border-bottom">
    <nav class="level">
      <div class="level-left">
        <div class="level-item">
          <DateSelector v-model="startDate" label="From" />
        </div>
      </div>
      <div class="level-left">
        <div class="level-item">
          <DateSelector v-model="endDate" label="To" />
        </div>
      </div>
      <p class="level-item"><button @click="loadData" class="button is-link">View</button></p>
    </nav>
  </div>
</template>
