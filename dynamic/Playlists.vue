<!-- Playlists.vue -->

<template>
    <div>
        <div class="container">
            <section class="search-header">
                <div id="search-box">
                    <input
                        name="q"
                        id="search-query"
                        placeholder="Search..."
                        autocapitalize="off"
                        autocomplete="off"
                        autocorrect="off"
                        spellcheck="false"
                        type="search"
                        v-model="query"
                        v-on:keyup.enter="search"
                    />
                </div>
            </section>
            <section class="section-search-results" v-if="matches">
                <div id="search-hits">
                    <table>
                        <thead>
                            <tr>
                                <th class="track" style="width: 25%;">Track</th>
                                <th class="artist" style="width: 25%;">Artist</th>
                                <th class="album" style="width: 25%;">Album</th>
                                <th class="album" style="width: 25%;">Playlists</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="track in matches" class="track">
                                <td class="track">{{ track.name }}</td>
                                <td class="artist">{{ track.artists | joinArtists }}</td>
                                <td class="album">{{ track.album }}</td>
                                <td class="playlists">
                                    <span v-for="pl in track.playlists">
                                        <a :href="'?id=' + pl.id" v-on:click="openPlaylist(pl.id, $event)">{{ pl.name }}</a>
                                        &nbsp;
                                    </span>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </section>
            <div v-masonry item-selector=".item" class="playlists">
                <div v-masonry-tile class="folder" v-for="folder in folders" v-if="folder.visible">
                    <h3>{{ folder.name }}</h3>
                    <ul>
                        <li
                            v-for="playlist in folder.playlists"
                            class="playlist"
                            :class="{ 'recently-modified': playlist.recentlyModified }"
                        >
                            <a :href="'?id=' + playlist.id" v-on:click="openPlaylist(playlist.id, $event)">{{ playlist.name }}</a>
                            <span class="details">
                                {{ playlist.numberOfTracks }} tracks, modified {{ playlist.lastModified | prettyTime }}
                            </span>
                        </li>
                    </ul>
                </div>
            </div>
        </div>
    </div>
</template>
<script>
import moment from 'moment'

import { Playlists } from './playlists'

export default {
    name: 'Home',
    props: {},
    data: () => {
        return {
            query: '',
            folders: [],
            matches: null,
        }
    },
    created() {
        this.refresh()
    },
    filters: {
        prettyTime(value) {
            return moment(value).format('MMM Do YYYY')
        },
        joinArtists(value) {
            return _(value).join(', ')
        },
        joinPlaylists(value) {
            return _(value)
                .map('name')
                .join(', ')
        },
    },
    methods: {
        refresh() {
            new Playlists().playlists().then(data => {
                this.folders = data.folders
            })
        },

        openPlaylist(id, ev) {
            ev.preventDefault()
            this.$emit('navigate', '?id=' + id)
        },

        search(ev) {
            new Playlists().search(this.query).then(data => {
                this.matches = data
            })
        },
    },
}
</script>
<style></style>
