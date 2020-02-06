<!-- Router.vue -->

<template>
    <div>
        {{ url }}
        <component :is="visible" :query="query" @navigate="handleNavigate"></component>
    </div>
</template>

<script>
import Playlists from './Playlists'
import Playlist from './Playlist'
import Router from './Router'

function parseQuery(queryString) {
    const query = {}
    const pairs = (queryString[0] === '?' ? queryString.substr(1) : queryString).split('&')
    for (let i = 0; i < pairs.length; i++) {
        const pair = pairs[i].split('=')
        query[decodeURIComponent(pair[0])] = decodeURIComponent(pair[1] || '')
    }
    return query
}

export default {
    data() {
        console.log('data', navigator)

        return {
            path: window.location.pathname,
            rawQuery: window.location.search,
        }
    },
    methods: {
        handleNavigate(url) {
            console.log('navigate', url)
            history.pushState({}, '', url)
            this.rawQuery = url
        },
    },
    computed: {
        query() {
            return parseQuery(this.rawQuery)
        },
        visible() {
            const query = parseQuery(this.rawQuery)
            if (query.id) {
                return Playlist
            }
            return Playlists
        },
    },
    created() {
        window.onpopstate = ev => {
            console.log('location: ' + window.location + ', state: ' + JSON.stringify(event.state))
            this.handleNavigate(window.location.search)
        }
    },
}
</script>
