
class Stats {
    constructor() {
        this.data = null;
    }

    fetchStats = async () => {
        try {
            const resp = await fetch('/stats');
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
    }
}

document.addEventListener('DOMContentLoaded', function () {
    const stats = new Stats();
    stats.fetchStats();
});