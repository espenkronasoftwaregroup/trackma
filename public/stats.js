
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
        console.log('constructor', this.groupPerHour);

        const url = new URL(document.location).searchParams;
        const d = luxon.DateTime.now();
        this.start = url.get('start') ?? d.minus({ days: 1}).toString().substring(0, 10);
        this.end = url.get('end') ?? d.toString().substring(0, 10);

        this.quickSyncGraph = new ApexCharts(document.getElementById('quicksyncs-per-hour'), {
            chart: {
                id: 'quicksyncs-per-hour',
                type: 'line',
                height: this.lineChartHeight
            },
            title: {
                text: 'Quick syncs per hour'
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
                text: 'Page views per hour'
            },
            series: [{
                name: 'Views',
                data: []
            }],
            xaxis: {
                categories: []
            }
        });

        this.quickSyncGraph.render();
        this.pageViewsGraph.render();

        document.getElementById('hourSwitch').onchange = e => {
            this.groupPerHour = !e.currentTarget.checked;
            console.log('callback', this.groupPerHour);
            this.updateRequestsPerHour();
            this.updateQuickSyncsPerHour();
        };

        if (window.location.href.includes("#quicksyncs")) {
            document.getElementById('quicksync-tab').classList.add('is-active');
            document.getElementById('quicksync-table').classList.remove('hidden');
            document.getElementById('pageview-tab').classList.remove('is-active');
            document.getElementById('pageview-table').classList.add('hidden');
        }

        window.addEventListener('popstate', () => {
            if (window.location.href.includes('#quicksyncs')) {
                document.getElementById('quicksync-tab').classList.add('is-active');
                document.getElementById('quicksync-table').classList.remove('hidden');
                document.getElementById('pageview-tab').classList.remove('is-active');
                document.getElementById('pageview-table').classList.add('hidden');

            } else {
                document.getElementById('quicksync-tab').classList.remove('is-active');
                document.getElementById('quicksync-table').classList.add('hidden');
                document.getElementById('pageview-tab').classList.add('is-active');
                document.getElementById('pageview-table').classList.remove('hidden');
            }
        });
    }

    fetchStats = async () => {
        try {
            const urlParams = new URLSearchParams();
            urlParams.set('start', this.start);
            urlParams.set('end', this.end);
            const resp = await fetch('/stats?' + urlParams.toString());
            this.data = await resp.json();
            console.log(this.data);

            this.updateData()
        } catch (err) {

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
        const timePoints = [];
        const hits = [];

        if (this.groupPerHour) {
            for (const [key, value] of Object.entries(this.data.quick_syncs_per_hour)) {
                timePoints.push(luxon.DateTime.fromFormat(key, 'yyyy-MM-dd HH', {zone: 'utc'}).toLocal().toFormat('HH'));
                hits.push(value);
            }
        } else {
            const groups = {};

            for (const key of Object.keys(this.data.quick_syncs_per_hour)) {
                const d = key.substring(0, 10);

                if (groups[d] === undefined) {
                    groups[d] = 0;
                }

                groups[d] += this.data.quick_syncs_per_hour[key];
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
                height: this.barChartHeight
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
        document.getElementById("current-visitors").textContent = this.numberFormatter.format(this.data.current_visitors);
        document.getElementById("total-page-views").textContent = this.numberFormatter.format(this.data.total_page_views);

        this.updateRequestsPerHour();
        this.updateQuickSyncsPerHour();
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