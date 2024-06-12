<script setup>
import { ref, computed } from 'vue';
import { DateTime } from 'luxon';
import { useStats } from '@/stores/stats.js';

const ss = useStats();
const activeTab = ref('page-views');

const views = computed(() => {
    const result = {
        series: [{
            name: 'Views',
            data: []
        }],
        opts: {
            title: {
                text: 'Page views per time'
            },
            xaxis: {
                categories: []
            }
        }
    }

    if (ss.groupPerHour) {
        for (const [key, value] of Object.entries(ss.stats.page_views_per_hour)) {
            result.opts.xaxis.categories.push(luxon.DateTime.fromFormat(key, 'yyyy-MM-dd HH', { zone: 'utc' }).toLocal().toFormat('HH'));
            result.series[0].data.push(value);
        }
    } else {
        const groups = {};

        for (const key of Object.keys(ss.stats.page_views_per_hour)) {
            const d = key.substring(0, 10);

            if (groups[d] === undefined) {
                groups[d] = 0;
            }

            groups[d] += ss.stats.page_views_per_hour[key];
        }

        for (const [key, value] of Object.entries(groups)) {
            result.opts.xaxis.categories.push(key);
            result.series[0].data.push(value);
        }
    }

    return result;
});

const syncs = computed(() => {
    const result = {
        series: [{
            name: 'Quick syncs',
            data: []
        }],
        opts: {
            title: {
                text: 'Quick syncs per time'
            },
            xaxis: {
                categories: []
            }
        }
    }

    if (ss.groupPerHour) {
        for (const [key, value] of Object.entries(ss.stats.events_per_name_and_hour.quick_sync)) {
            result.opts.xaxis.categories.push(luxon.DateTime.fromFormat(key, 'yyyy-MM-dd HH', { zone: 'utc' }).toLocal().toFormat('HH'));
            result.series[0].data.push(value);
        }
    } else {
        const groups = {};

        for (const key of Object.keys(ss.stats.events_per_name_and_hour.quick_sync)) {
            const d = key.substring(0, 10);

            if (groups[d] === undefined) {
                groups[d] = 0;
            }

            groups[d] += ss.stats.events_per_name_and_hour.quick_sync[key];
        }

        for (const [key, value] of Object.entries(groups)) {
            result.opts.xaxis.categories.push(key);
            result.series[0].data.push(value);
        }
    }

    return result;
});

const accounts = computed(() => {
    const result = {
        series: [{
            name: 'Accounts created',
            data: []
        }],
        opts: {
            title: {
                text: 'Account creations per time'
            },
            xaxis: {
                categories: []
            }
        }
    }

    if (ss.groupPerHour) {
        for (const [key, value] of Object.entries(ss.stats.events_per_name_and_hour.account_created)) {
            result.opts.xaxis.categories.push(luxon.DateTime.fromFormat(key, 'yyyy-MM-dd HH', { zone: 'utc' }).toLocal().toFormat('HH'));
            result.series[0].data.push(value);
        }
    } else {
        const groups = {};

        for (const key of Object.keys(ss.stats.events_per_name_and_hour.account_created)) {
            const d = key.substring(0, 10);

            if (groups[d] === undefined) {
                groups[d] = 0;
            }

            groups[d] += ss.stats.events_per_name_and_hour.account_created[key];
        }

        for (const [key, value] of Object.entries(groups)) {
            result.opts.xaxis.categories.push(key);
            result.series[0].data.push(value);
        }
    }

    return result;
});

</script>

<template>
    <div class="section">
        <div class="tabs">
            <ul>
                <li @click="activeTab = 'page-views'" :class="{ 'is-active': (activeTab === 'page-views') }">
                    <a>Page views</a>
                </li>
                <li @click="activeTab = 'quick-syncs'" :class="{ 'is-active': (activeTab === 'quick-syncs') }">
                    <a>Quick syncs</a>
                </li>
                <li @click="activeTab = 'account-creations'"
                    :class="{ 'is-active': (activeTab === 'account-creations') }">
                    <a>Account creations</a>
                </li>
            </ul>
        </div>
        <div v-if="activeTab === 'page-views'">
            <apexchart type="line" height="400" :options="views.opts" :series="views.series"></apexchart>
        </div>
        <div v-if="activeTab === 'quick-syncs'">
            <apexchart type="line" height="400" :options="syncs.opts" :series="syncs.series"></apexchart>
        </div>
        <div v-if="activeTab === 'account-creations'">
            <apexchart type="line" height="400" :options="accounts.opts" :series="accounts.series"></apexchart>
        </div>
    </div>
</template>