
function getFlagEmoji(countryCode) {
    const codePoints = countryCode
        .toUpperCase()
        .split('')
        .map(char =>  127397 + char.charCodeAt());
    return String.fromCodePoint(...codePoints);
}

class Stats {
    constructor() {
        this.numberFormatter = new Intl.NumberFormat('sv-SE');
        this.barChartHeight = '300px';
        this.lineChartHeight = '400px';
        this.data = null;
        this.groupPerHour = !document.getElementById('hourSwitch').checked;

        const url = new URL(document.location).searchParams;
        const d = luxon.DateTime.now();
        this.start = url.get('start') ?? d.minus({ days: 1}).toString().substring(0, 10);
        this.end = url.get('end') ?? d.toString().substring(0, 10);
        document.getElementById('start-date').value = this.start;
        document.getElementById('end-date').value = this.end;

        this.quickSyncGraph = new ApexCharts(document.getElementById('quicksyncs-per-hour'), {
            chart: {
                id: 'quicksyncs-per-hour',
                type: 'line',
                height: this.lineChartHeight
            },
            title: {
                text: 'Quick syncs per time'
            },
            series: [{
                name: 'Syncs',
                data: []
            }],
            xaxis: {
                categories: []
            }
        });

        this.pageViewsGraph = new ApexCharts(document.getElementById('pageviews-per-hour'), {
            chart: {
                id: 'pageviews-per-hour',
                type: 'line',
                height: this.lineChartHeight
            },
            title: {
                text: 'Page views per time'
            },
            series: [{
                name: 'Views',
                data: []
            }],
            xaxis: {
                categories: []
            }
        });

        this.accountCreationGraph = new ApexCharts(document.getElementById('account-creations-per-hour'), {
            chart: {
                id: 'account-creations',
                type: 'line',
                height: this.lineChartHeight
            },
            title: {
                text: 'Account creations per time'
            },
            series: [{
                name: 'Creations',
                data: []
            }],
            xaxis: {
                categories: []
            }
        });

        this.quickSyncGraph.render();
        this.pageViewsGraph.render();
        this.accountCreationGraph.render();

        document.getElementById('hourSwitch').onchange = e => {
            this.groupPerHour = !e.currentTarget.checked;
            this.updateRequestsPerHour();
            this.updateQuickSyncsPerHour();
            this.updateAccountCreations();
        };

        if (window.location.href.includes('#')) {
            this.setActiveTab();
        }

        window.addEventListener('popstate', this.setActiveTab);
    }

    setActiveTab = () => {
        const anchor = window.location.hash.substring(1);
        for (const element of document.querySelectorAll('div.tabs ul li')) {
            if (element.id === anchor) {
                element.classList.add('is-active');
                document.getElementById(element.getAttribute('data-table-element')).classList.remove('hidden');
            } else {
                element.classList.remove('is-active');
                document.getElementById(element.getAttribute('data-table-element')).classList.add('hidden');
            }
        }
    }

    fetchStats = async () => {
        try {
            const urlParams = new URLSearchParams();
            urlParams.set('start', this.start);
            urlParams.set('end', this.end);
            const resp = await fetch('/stats?' + urlParams.toString());
            this.data = await resp.json();
            this.updateData()
        } catch (err) {
            console.error(err);
        }
    }

    updateRequestsPerHour = () => {
        const timePoints = [];
        const hits = [];

        if (this.groupPerHour) {
            for (const [key, value] of Object.entries(this.data.page_views_per_hour)) {
                timePoints.push(luxon.DateTime.fromFormat(key, 'yyyy-MM-dd HH', {zone: 'utc'}).toLocal().toFormat('HH'));
                hits.push(value);
            }
        } else {
            const groups = {};

            for (const key of Object.keys(this.data.page_views_per_hour)) {
                const d = key.substring(0, 10);

                if (groups[d] === undefined) {
                    groups[d] = 0;
                }

                groups[d] += this.data.page_views_per_hour[key];
            }

            for (const [key, value] of Object.entries(groups)) {
                timePoints.push(key);
                hits.push(value);
            }
        }

        ApexCharts.exec('pageviews-per-hour', 'updateOptions', {
            series: [{
                name: 'Views',
                data: hits
            }],
            xaxis: {
                categories: timePoints
            }
        });
    }

    updateQuickSyncsPerHour = () => {
        if (!this.data.events_per_name_and_hour?.quick_sync) return;

        const timePoints = [];
        const hits = [];

        if (this.groupPerHour) {
            for (const [key, value] of Object.entries(this.data.events_per_name_and_hour.quick_sync)) {
                timePoints.push(luxon.DateTime.fromFormat(key, 'yyyy-MM-dd HH', {zone: 'utc'}).toLocal().toFormat('HH'));
                hits.push(value);
            }
        } else {
            const groups = {};

            for (const key of Object.keys(this.data.events_per_name_and_hour.quick_sync)) {
                const d = key.substring(0, 10);

                if (groups[d] === undefined) {
                    groups[d] = 0;
                }

                groups[d] += this.data.events_per_name_and_hour.quick_sync[key];
            }

            for (const [key, value] of Object.entries(groups)) {
                timePoints.push(key);
                hits.push(value);
            }
        }

        ApexCharts.exec('quicksyncs-per-hour', 'updateOptions', {
            series: [{
                name: 'Syncs',
                data: hits
            }],
            xaxis: {
                categories: timePoints
            }
        });
    }

    updateAccountCreations = () => {
        if (!this.data.events_per_name_and_hour?.account_created) return;

        const timePoints = [];
        const hits = [];

        if (this.groupPerHour) {
            for (const [key, value] of Object.entries(this.data.events_per_name_and_hour.account_created)) {
                timePoints.push(luxon.DateTime.fromFormat(key, 'yyyy-MM-dd HH', {zone: 'utc'}).toLocal().toFormat('HH'));
                hits.push(value);
            }
        } else {
            const groups = {};

            for (const key of Object.keys(this.data.events_per_name_and_hour.account_created)) {
                const d = key.substring(0, 10);

                if (groups[d] === undefined) {
                    groups[d] = 0;
                }

                groups[d] += this.data.events_per_name_and_hour.account_created[key];
            }

            for (const [key, value] of Object.entries(groups)) {
                timePoints.push(key);
                hits.push(value);
            }
        }

        ApexCharts.exec('account-creations', 'updateOptions', {
            series: [{
                name: 'Creations',
                data: hits
            }],
            xaxis: {
                categories: timePoints
            }
        });
    }

    updateRequestsPerIp = () => {
        let sortable = [...this.data.requests_per_ip];
        sortable.sort((a, b) => b.count - a.count);
        sortable = sortable.slice(0, 10)

        const data = [];
        for (const pair of sortable) {
            data.push({
                x: `${pair.ip} - ${getFlagEmoji(pair.country)}`,
                y: pair.count
            })
        }

        new ApexCharts(document.getElementById('requests-per-ip'), {
            chart: {
                id: 'requests-per-ip',
                type: 'bar',
                height: this.barChartHeight,
                events: {
                    click: (e, chartContext, config) => {
                        console.log(config)
                    }
                }
            },
            plotOptions: {
                bar: {
                    horizontal: true
                }
            },
            title: {
                text: 'Top 10 IPs with most requests'
            },
            series: [{
                data
            }],
        }).render();
    }

    updateReferrers = () => {
        let sortable = Object.entries(this.data.referrers)
        sortable.sort((a, b) => b[1] - a[1]);
        sortable = sortable.slice(0, 10)

        const data = [];
        for (const pair of sortable) {
            data.push({
                x: pair[0],
                y: pair[1]
            })
        }

        new ApexCharts(document.getElementById('referrers'), {
            chart: {
                id: 'referrers',
                type: 'bar',
                height: this.barChartHeight
            },
            plotOptions: {
                bar: {
                    horizontal: true
                }
            },
            title: {
                text: 'Top 10 referral sites'
            },
            series: [{
                data
            }],
        }).render();
    }

    updateVisitorsPerCountry = () => {
        let sortable = Object.entries(this.data.visitors_per_country)
        sortable.sort((a, b) => b[1] - a[1]);
        sortable = sortable.slice(0, 10)
        const regionNames = new Intl.DisplayNames(['en'], {type: 'region'});

        const data = [];
        for (const pair of sortable) {
            data.push({
                x: `${regionNames.of(pair[0])} - ${getFlagEmoji(pair[0])}`,
                y: pair[1]
            })
        }

        new ApexCharts(document.getElementById('visitors-per-country'), {
            chart: {
                id: 'visitors-per-country',
                type: 'bar',
                height: this.barChartHeight
            },
            plotOptions: {
                bar: {
                    horizontal: true
                }
            },
            title: {
                text: 'Top 10 countries with most visitors'
            },
            series: [{
                data
            }],
        }).render();
    }

    updateVisitorsPerUtmSource = () => {
        let sortable = Object.entries(this.data.visitors_per_utm_source)
        sortable.sort((a, b) => b[1] - a[1]);
        sortable = sortable.slice(0, 10)

        const data = [];
        for (const pair of sortable) {
            data.push({
                x: pair[0],
                y: pair[1]
            })
        }

        new ApexCharts(document.getElementById('visitors-per-utm-source'), {
            chart: {
                id: 'visitors-per-utm-source',
                type: 'bar',
                height: this.barChartHeight
            },
            plotOptions: {
                bar: {
                    horizontal: true
                }
            },
            title: {
                text: 'Top 10 utm sources'
            },
            series: [{
                data
            }],
        }).render();
    }

    updateData = () => {
        document.getElementById("total-visitors").textContent = this.numberFormatter.format(this.data.total_visitors);
        document.getElementById("total-page-views").textContent = this.numberFormatter.format(this.data.total_page_views);
        document.getElementById('subscriptions-started').textContent = this.numberFormatter.format(this.data.subscriptions_started);
        document.getElementById('orders-completed').textContent = this.numberFormatter.format(this.data.orders_completed);
        document.getElementById('trials-started').textContent = this.numberFormatter.format(this.data.trials_started);
        document.getElementById('accounts-created').textContent = this.numberFormatter.format(this.data.accounts_created);

        this.updateRequestsPerHour();
        this.updateQuickSyncsPerHour();
        this.updateAccountCreations();
        this.updateRequestsPerIp();
        this.updateVisitorsPerCountry();
        this.updateReferrers();
        this.updateVisitorsPerUtmSource();

        document.getElementById('spinner').classList.add('hidden');
        document.getElementById('hider').classList.remove('hidden');
    }
}

document.addEventListener('DOMContentLoaded', function () {
    const stats = new Stats();
    stats.fetchStats();
});