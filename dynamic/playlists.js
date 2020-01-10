import _ from 'lodash'
import moment from 'moment'

export class Playlists {
	constructor() {
	}

	playlists() {
		return Promise.all([
			this._json("/music/folders.json"),
			this._json("/music/playlists.json"),
		]).then(data => {
			console.log(data)
			return {
				folders: data[0].folders,
				playlists: data[1].playlists,
				sorted: this.sortIntoFolders(data[0].folders, data[1].playlists),
			};
		}).then(data => {
			console.log(data)
			return data.sorted
		});
	}

	playlist(id) {
		return Promise.all([
			this._json("/music/folders.json"),
			this._json("/music/playlists.json"),
			this._json("/music/playlist-" + id + ".json"),
		]).then(data => {
			return {
				folders: data[0].folders,
				tracks: data[2],
				playlist: _(data[1].playlists).filter(pl => pl.id == id).first()
			}
		});
	}

	_json(url) {
		return fetch(url).then(response => response.json())
	}

	getFirstMatchingFolder(folders, playlist) {
		return _.extend(playlist, {
			folder: _(folders).filter(folder => {
				// This should be a flag on the playlist, or at least move jlewalle out.
				if (folder.subscribed === true && playlist.owner.id != 'jlewalle') {
					return true;
				}
				return _(folder.pattern).some(pattern => {
					return playlist.name.match(pattern)
				});
			}).first().name
		});
	}

	decoratePlaylist(playlist) {
		return _.merge(playlist, {
			recentlyModified: playlist.lastModified && moment(playlist.lastModified).isAfter(moment().subtract(7, 'days')),
		});
	}

	sortFolderPlaylists(folder, playlists) {
		if (folder.sorted) {
			const sortedPlaylists = _.sortBy(playlists || [], p => {
				const match = _(folder.pattern).map(pattern => {
					return p.name.match(pattern)
				}).first();
				const asNumber = Number(match[1])
				if (isNaN(asNumber)) {
					return match[1]
				}
				return asNumber
			});
			if (folder.reversed === true) {
				return _.reverse(sortedPlaylists)
			}
			return sortedPlaylists
		}
		return  _.reverse(_.sortBy(playlists || [], p => moment(p.lastModified)))
	}

	sortIntoFolders(folders, playlists) {
		if (!_.some(folders) || !_.some(playlists)) {
			return { folders: [] }
		}

		const withFolder = _(playlists).map(p => this.decoratePlaylist(p)).filter(p => p.name != 'Discover Weekly').map(p => this.getFirstMatchingFolder(folders, p))
		const byFolder = withFolder.groupBy('folder').value()

		const foldersWithPlaylists = _(folders).map(config => {
			return {
				subscribed: config.subscribed,
				order: config.order,
				name: config.name,
				reduced: true,
				playlists: this.sortFolderPlaylists(config, byFolder[config.name])
			};
		}).orderBy('order').filter(f => !f.subscribed).value()

		return {
			folders: foldersWithPlaylists
		};
	}

}
