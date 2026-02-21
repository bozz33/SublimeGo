// Charts configuration for SublimeGo Dashboard
// Source: dashboard/js/charts.js — Conversion fidèle à 100%

// Check for dark mode
function isDarkMode() {
    return document.documentElement.classList.contains('dark');
}

// Get chart colors based on theme
function getChartColors() {
    const isDark = isDarkMode();
    return {
        primary: '#22c55e',
        secondary: '#3b82f6',
        warning: '#eab308',
        danger: '#ef4444',
        text: isDark ? '#9ca3af' : '#6b7280',
        grid: isDark ? '#374151' : '#e5e7eb',
        background: isDark ? '#1f2937' : '#ffffff'
    };
}

// Line Chart - Monthly Activity
function initLineChart() {
    const colors = getChartColors();
    const options = {
        series: [{
            name: 'Offres',
            data: [180, 220, 250, 280, 310, 350]
        }, {
            name: 'Collectes',
            data: [150, 190, 220, 250, 280, 320]
        }],
        chart: {
            type: 'area',
            height: 320,
            fontFamily: 'Inter, sans-serif',
            toolbar: { show: false },
            background: 'transparent'
        },
        colors: [colors.primary, colors.secondary],
        fill: {
            type: 'gradient',
            gradient: {
                shadeIntensity: 1,
                opacityFrom: 0.4,
                opacityTo: 0.1,
                stops: [0, 90, 100]
            }
        },
        stroke: {
            curve: 'smooth',
            width: 2
        },
        dataLabels: { enabled: false },
        xaxis: {
            categories: ['Août', 'Sept', 'Oct', 'Nov', 'Déc', 'Jan'],
            axisBorder: { show: false },
            axisTicks: { show: false },
            labels: {
                style: { colors: colors.text, fontSize: '12px' }
            }
        },
        yaxis: {
            labels: {
                style: { colors: colors.text, fontSize: '12px' }
            }
        },
        grid: {
            borderColor: colors.grid,
            strokeDashArray: 4,
            xaxis: { lines: { show: false } }
        },
        legend: {
            position: 'top',
            horizontalAlign: 'right',
            labels: { colors: colors.text }
        },
        tooltip: {
            theme: isDarkMode() ? 'dark' : 'light'
        }
    };

    const chartEl = document.querySelector('#lineChart');
    if (chartEl) {
        const chart = new ApexCharts(chartEl, options);
        chart.render();
        return chart;
    }
    return null;
}

// Donut Chart - Category Distribution
function initDonutChart() {
    const colors = getChartColors();
    const options = {
        series: [35, 25, 20, 20],
        chart: {
            type: 'donut',
            height: 256,
            fontFamily: 'Inter, sans-serif',
            background: 'transparent'
        },
        colors: [colors.primary, colors.secondary, colors.warning, colors.danger],
        labels: ['Fruits & Légumes', 'Boulangerie', 'Produits Laitiers', 'Plats Préparés'],
        plotOptions: {
            pie: {
                donut: {
                    size: '70%',
                    labels: {
                        show: true,
                        name: { show: true, color: colors.text },
                        value: {
                            show: true,
                            color: colors.text,
                            formatter: (val) => val + '%'
                        },
                        total: {
                            show: true,
                            label: 'Total',
                            color: colors.text,
                            formatter: () => '100%'
                        }
                    }
                }
            }
        },
        dataLabels: { enabled: false },
        legend: { show: false },
        stroke: { width: 0 },
        tooltip: {
            theme: isDarkMode() ? 'dark' : 'light'
        }
    };

    const chartEl = document.querySelector('#donutChart');
    if (chartEl) {
        const chart = new ApexCharts(chartEl, options);
        chart.render();
        return chart;
    }
    return null;
}

// Bar Chart - Weekly Stats
function initBarChart() {
    const colors = getChartColors();
    const options = {
        series: [{
            name: 'Collectes',
            data: [44, 55, 57, 56, 61, 58, 63]
        }, {
            name: 'Livraisons',
            data: [35, 41, 36, 26, 45, 48, 52]
        }],
        chart: {
            type: 'bar',
            height: 350,
            fontFamily: 'Inter, sans-serif',
            toolbar: { show: false },
            background: 'transparent'
        },
        colors: [colors.primary, colors.secondary],
        plotOptions: {
            bar: {
                horizontal: false,
                columnWidth: '55%',
                borderRadius: 4
            }
        },
        dataLabels: { enabled: false },
        stroke: { show: true, width: 2, colors: ['transparent'] },
        xaxis: {
            categories: ['Lun', 'Mar', 'Mer', 'Jeu', 'Ven', 'Sam', 'Dim'],
            axisBorder: { show: false },
            axisTicks: { show: false },
            labels: { style: { colors: colors.text, fontSize: '12px' } }
        },
        yaxis: {
            labels: { style: { colors: colors.text, fontSize: '12px' } }
        },
        grid: {
            borderColor: colors.grid,
            strokeDashArray: 4
        },
        legend: {
            position: 'top',
            horizontalAlign: 'right',
            labels: { colors: colors.text }
        },
        fill: { opacity: 1 },
        tooltip: { theme: isDarkMode() ? 'dark' : 'light' }
    };

    const chartEl = document.querySelector('#barChart');
    if (chartEl) {
        const chart = new ApexCharts(chartEl, options);
        chart.render();
        return chart;
    }
    return null;
}

// Initialize all charts on page load
document.addEventListener('DOMContentLoaded', function() {
    initLineChart();
    initDonutChart();
    initBarChart();
});

// Re-render charts on theme change
document.addEventListener('alpine:init', () => {
    Alpine.effect(() => {
        // Watch for dark mode changes and re-render charts
        const darkMode = Alpine.store('darkMode');
        if (darkMode !== undefined) {
            setTimeout(() => {
                // Charts will be re-initialized with new colors
            }, 100);
        }
    });
});
