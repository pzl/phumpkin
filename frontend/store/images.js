import Vue from 'vue'

export const state = () => ({
	images:[],
	dirs: [],
	selected: [],
	loading: false,
	err: false,
	sort: 'name',
	sort_asc: true,
	path: [],
})

export const mutations = {
	startLoading (state) { state.loading = true },
	stopLoading (state) { state.loading = false },
	errorLoading(state) { state.err = true },
	clearErr(state) { state.err = false },
	addImages(state, images) { state.images.push(...images) },
	setImages(state,images) { state.images = images },
	setDirs(state, dirs) { state.dirs = dirs },
	clearPath(state) { state.path = [] },
	pushPath(state, dir) { state.path.push(dir) },
	popPath(state) { state.path.pop() },
	clearImages(state) { state.images = [] },
	clearSelection (state) { state.selected = [] },
	select (state, image) { state.selected.push(image) },
	unselect (state, image) {
		const idx = state.selected.indexOf(image)
		if (idx === -1) {
			return
		}
		state.selected.splice(idx, 1)
	},
	rate(state, { image, rating }) {
		state.images[image].xmp.rating = rating
	},
	sortBy(state, by) { state.sort = by },
	sortDir(state, dir){ state.sort_asc = dir },
}

export const actions = {
	loadImages({ commit, state }) {
		commit('startLoading')
		commit('clearErr')

		let p;

		if (this.$sock.connected()) {
			p = this.$sock.send({
					action:"list",
					params: {
						offset: state.images.length,
						count: 10,
						sort: state.sort,
						sort_dir: state.sort_asc ? 'asc' : 'desc',
						path: state.path.join('/'),
					}
				})
		} else {
			// fall back to HTTP
			let server = location.origin
			if (server === "http://localhost:3000") {
				// @todo: remove local dev hack
				server = "http://localhost:6001"
			}
			p = this.$axios.$get(
				server+"/api/v1/photos?"+
					"count=30&"+
					"offset="+(state.images.length||0)+"&"+
					"sort="+state.sort+"&"+
					"sort_dir="+(state.sort_asc ? 'asc' : 'desc')+"&"+
					"path="+state.path.join('/')
			)
		}
		p.then(data => {
			commit('addImages', data.photos)
			commit('setDirs', data.dirs)
		})
		.catch(error => {
			console.log('load image error: ')
			console.log(error)
			commit('errorLoading')
		})
		.finally(() => {
			commit('stopLoading')
		})
		return p
	},
	toggleSelect({ commit, state }, image) {
		if (state.selected.includes(image)) {
			commit('unselect', image)
		} else {
			commit('select', image)
		}
	},
	setSelection({ commit }, image) {
		commit('clearSelection')
		commit('select', image)
	},
	addSelection({ commit, state }, image) {
		if (!state.selected.includes(image)) {
			commit('select', image)
		}
	},
	resetImages({ commit }) {
		commit('clearSelection')
		commit('clearImages')
	},
}