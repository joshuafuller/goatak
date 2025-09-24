const app = Vue.createApp({
    data: function () {
        return {
            data: [],
            current: null,
            ts: 0,
        }
    },

    mounted() {
        this.renew();
        setInterval(this.renew, 60000);
    },
    computed: {
        all: function () {
            return this.ts && this.data;
        },
    },
    methods: {
        renew: function () {
            let vm = this;

            fetch('/api/file', {redirect: 'manual'})
                .then(resp => {
                    if (!resp.ok) {
                        window.location.reload();
                    }
                    return resp.json();
                })
                .then(data => {
                    vm.data = data;
                    vm.ts += 1;
                });
        },
        printCoords: function (lat, lng) {
            return lat.toFixed(6) + "," + lng.toFixed(6);
        },
        dt: dtShort,
    },
});

app.mount('#app');