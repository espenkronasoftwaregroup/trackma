
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

    updateData = () => {
        document.getElementById("total-visitors").textContent = this.data.total_visitors;
        document.getElementById("current-visitors").textContent = this.data.current_visitors;
        document.getElementById("total-page-views").textContent = this.data.total_page_views;

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
                data: [2, 33, 14, 8]
            }]
        }).render();
    }
}

document.addEventListener('DOMContentLoaded', function () {
    const stats = new Stats();
    stats.fetchStats();
});