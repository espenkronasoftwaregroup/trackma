<script setup>
import DateSelector from '@/components/DateSelector.vue';
import NumberStat from '@/components/NumberStat.vue';
import LineCharts from '@/components/LineCharts.vue';
import BarCharts from '@/components/BarCharts.vue'
import { useStats } from '@/stores/stats.js';
import { ref } from 'vue';

const ss = useStats();

const startDate = ref(new Date().toISOString().substring(0,10));
const endDate = ref(new Date().toISOString().substring(0, 10));

async function loadData() {
  ss.fetchStats(startDate.value, endDate.value);
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
  <div class="section numbers" v-if="ss.stats">
    <NumberStat :value="ss.stats.total_visitors" label="Visitors" />
    <NumberStat :value="ss.stats.total_page_views" label="Page views" />
    <NumberStat :value="ss.stats.accounts_created" label="Accounts created" />
    <NumberStat :value="ss.stats.orders_completed" label="Orders completed" />
    <NumberStat :value="ss.stats.subscriptions_started" label="Subscriptons started" />
    <NumberStat :value="ss.stats.trials_started" label="Trials started" />
  </div>
  <LineCharts v-if="ss.stats" />
  <BarCharts v-if="ss.stats" />
</template>
