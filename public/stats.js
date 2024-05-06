
class Stats {
    constructor() {
        this.data = null;
    }

    fetchStats = async () => {
        try {
            const resp = await fetch('/stats');
            this.data = await resp.json();
        } catch (err) {

        }
    }
}

document.addEventListener('DOMContentLoaded', function () {
    const stats = new Stats();
    stats.fetchStats();
});