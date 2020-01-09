import Vue from 'vue'

import { VueMasonryPlugin } from 'vue-masonry'

import Router from './Router'

Vue.use(VueMasonryPlugin)

new Vue({
	render: h => h(Router)
}).$mount("#application");
