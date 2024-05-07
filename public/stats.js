
class Stats {
    constructor() {
        this.data = null;

        const url = new URL(document.location).searchParams;
        const d = luxon.DateTime.now();
        this.start = url.get('start') ?? d.minus({ days: 1}).toString().substring(0, 10);
        this.end = url.get('end') ?? d.toString().substring(0, 10);
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
        const hours = [];
        const hits = [];

        for (const [key, value] of Object.entries(this.data.requests_per_hour)) {
            hours.push(luxon.DateTime.fromFormat(key, 'yyyy-MM-dd HH', { zone: 'utc' }).toLocal().toFormat('HH'));
            hits.push(value);
        }

        new ApexCharts(document.getElementById('requests-per-hour'), {
            chart: {
                id: 'mychart',
                type: 'line',
                height: '400px'
            },
            title: {
                text: 'Requests per hour'
            },
            series: [{
                name: 'Requests',
                data: hits
            }],
            xaxis: {
                categories: hours
            }
        }).render();
    }

    updateRequestsPerIp = () => {
        let sortable = Object.entries(this.data.requests_per_ip);
        sortable.sort((a, b) => b[1] - a[1]);
        sortable = sortable.slice(0, 10)

        const data = [];
        for (const pair of sortable) {
            data.push({
                x: pair[0],
                y: pair[1]
            })
        }

        new ApexCharts(document.getElementById('requests-per-ip'), {
            chart: {
                id: 'requests-per-ip',
                type: 'bar',
                height: '250px'
            },
            plotOptions: {
                bar: {
                    horizontal: true
                }
            },
            title: {
                text: 'Top 10 ips with most requests'
            },
            series: [{
                data
            }],
        }).render();
    }

    updateData = () => {
        document.getElementById("total-visitors").textContent = this.data.total_visitors;
        document.getElementById("current-visitors").textContent = this.data.current_visitors;
        document.getElementById("total-page-views").textContent = this.data.total_page_views;

        this.updateRequestsPerHour();
        this.updateRequestsPerIp();
    }
}

document.addEventListener('DOMContentLoaded', function () {
    const stats = new Stats();
    stats.fetchStats();
});