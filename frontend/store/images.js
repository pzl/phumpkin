import Vue from 'vue'

export const state = () => ({
	images:[],
	selected: [],
	loading: false,
})

export const mutations = {
	startLoading (state) { state.loading = true },
	stopLoading (state) { state.loading = false },
	setImages(state, images) { state.images = images },
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
	}
}

export const actions = {
	async loadImages({ commit }) {
		commit('startLoading')

		if (this.$sock.connected()) {
			this.$sock.send({action:"list"})
				.then(data => {
					commit('setImages', data)
					commit('clearSelection')
				})
				.catch(error => {
					console.log('error: ')
					console.log(error)
				})
				.finally(() => {
					commit('stopLoading')
				})
		} else {
			// fall back to HTTP
			try {
				// @todo: replace with location.origin + "/api.."
				const response = await this.$axios.$get("http://localhost:6001/api/v1/photos")
				commit('setImages', response.photos)
				commit('clearSelection')
			} catch (error) {
				// oh no
				console.log("error: ")
				console.log(error)
			}
			commit('stopLoading')
		}
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
	}
}