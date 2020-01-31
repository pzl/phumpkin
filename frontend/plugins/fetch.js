export default ({ app, store }, inject) => {
	const server = location.origin !== "http://localhost:3000" ? location.origin : "http://localhost:6001"

	const f = (url, data) => {
		const firstJoin = url.includes('?') ? '&' : '?'
		const q = ["count=30", 'offset=' + (store.state.images.images.length || 0)]

		if (store.state.images.sort) {
			q.push("sort=" + store.state.images.sortables[store.state.images.sort].text)
			q.push('sort_dir=' + (store.state.images.sort_asc ? 'asc' : 'desc'))
		}

		const path = url + firstJoin + q.join('&')
		if (data !== undefined) {
			return app.$axios.post(server + path, data)
		}
		return app.$axios.get(server + path)
	}

	inject('fetch', f)
}
