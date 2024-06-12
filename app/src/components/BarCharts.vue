<script setup>
import { computed } from 'vue';
import { DateTime } from 'luxon';
import { useStats } from '@/stores/stats.js';

const ss = useStats();

function getFlagEmoji(countryCode) {
    const codePoints = countryCode
        .toUpperCase()
        .split('')
        .map(char => 127397 + char.charCodeAt());
    return String.fromCodePoint(...codePoints);
}


const visitorsPerCountryOpts = {
    chart: {
        type: 'bar',
    },
    plotOptions: {
        bar: {
            horizontal: true
        }
    },
    title: {
        text: 'Top 10 countries with most visitors'
    },
};
const requestsPerCountryOpts = {
    chart: {
        type: 'bar',
    },
    plotOptions: {
        bar: {
            horizontal: true
        }
    },
    title: {
        text: 'Top 10 IPs with most requests'
    },
};
const referralOpts = {
    chart: {
        type: 'bar',
    },
    plotOptions: {
        bar: {
            horizontal: true
        }
    },
    title: {
        text: 'Top 10 referral sites'
    },
}
const utmOpts = {
    chart: {
        type: 'bar',
    },
    plotOptions: {
        bar: {
            horizontal: true
        }
    },
    title: {
        text: 'Top 10 utm sources'
    },
}

const visitorsPerCountry = computed(() => {
    let sortable = Object.entries(ss.stats.visitors_per_country)
    sortable.sort((a, b) => b[1] - a[1]);
    sortable = sortable.slice(0, 10)
    const regionNames = new Intl.DisplayNames(['en'], { type: 'region' });

    const data = [];
    for (const pair of sortable) {
        data.push({
            x: `${regionNames.of(pair[0])} - ${getFlagEmoji(pair[0])}`,
            y: pair[1]
        })
    }

    return [{ data }];
});

const requestsPerIp = computed(() => {
    let sortable = [...ss.stats.requests_per_ip];
    sortable.sort((a, b) => b.count - a.count);
    sortable = sortable.slice(0, 10)

    const data = [];
    for (const pair of sortable) {
        data.push({
            x: `${pair.ip} - ${getFlagEmoji(pair.country)}`,
            y: pair.count
        })
    }

    return [{ data }];
});

const referrals = computed(() => {
    let sortable = Object.entries(ss.stats.referrers)
    sortable.sort((a, b) => b[1] - a[1]);
    sortable = sortable.slice(0, 10)

    const data = [];
    for (const pair of sortable) {
        data.push({
            x: pair[0],
            y: pair[1]
        })
    }

    return [{ data }];
});

const utms = computed(() => {
    let sortable = Object.entries(ss.stats.visitors_per_utm_source)
    sortable.sort((a, b) => b[1] - a[1]);
    sortable = sortable.slice(0, 10)

    const data = [];
    for (const pair of sortable) {
        data.push({
            x: pair[0],
            y: pair[1]
        })
    }

    return [{ data }];
});

</script>

<template>
    <div class="section">
        <div class="columns">
            <div class="column">
                <apexchart type="bar" height="300" :options="requestsPerCountryOpts" :series="requestsPerIp" />
            </div>
            <div class="column">
                <apexchart type="bar" height="300" :options="visitorsPerCountryOpts" :series="visitorsPerCountry" />
            </div>
        </div>
    </div>
    <div class="section">
        <div class="columns">
            <div class="column">
                <apexchart type="bar" height="300" :options="referralOpts" :series="referrals" />
            </div>
            <div class="column">
                <apexchart type="bar" height="300" :options="utmOpts" :series="utms" />
            </div>
        </div>
    </div>
</template>