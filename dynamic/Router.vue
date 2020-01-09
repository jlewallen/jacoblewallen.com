<!-- Router.vue -->

<template>
<component :is="visible" :query="query"></component>
</template>

<script>
import Playlists from "./Playlists"
import Playlist from "./Playlist"
import Router from "./Router"

function test() {
}

export default {
	data() {
		function parseQuery(queryString) {
			const query = {}
			const pairs = (queryString[0] === '?' ? queryString.substr(1) : queryString).split('&')
			for (let i = 0; i < pairs.length; i++) {
				const pair = pairs[i].split('=')
				query[decodeURIComponent(pair[0])] = decodeURIComponent(pair[1] || '')
			}
			return query
		}

		return {
			path: window.location.pathname,
			query: parseQuery(window.location.search)
		};
	},
	computed: {
		visible() {
			if (this.query.id) {
				return Playlist
			}
			return Playlists
		}
	}
};
</script>
