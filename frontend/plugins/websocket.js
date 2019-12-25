export default ({ app, store }, inject) => {
	const pending = {}
	let redies = []
	let nextID = 1
	let init_connect = false


	const onopen = event => {
		store.commit('socket/setConnected')
		console.log("socket opened")
		console.log(event)

		event.target.onmessage = onmessage
		event.target.onclose = onclose
		event.target.onerror = onerror
		for (const r of redies) {
			r.resolve()
		}
		redies = []
	}
	const onmessage = event => {
		//console.log("got message")
		//console.log(event)

		const j = JSON.parse(event.data)
		if ("_id" in j && j._id in pending) {
			if ('error' in j) {
				pending[j._id].reject(JSON.parse(event.data).error)
			} else {
				pending[j._id].resolve(JSON.parse(event.data).data)
			}
			delete pending[j._id]
		}
	}
	const onclose = event => {
		store.commit('socket/setDisconnected')
		console.log('socket closed')
		console.log(event)
	}
	const onerror = event => {
		console.log("socket error")
		console.log(event)
		for (const r of redies) {
			r.reject()
		}
		redies = []
	}

	const t = {
		sock: null,
		reconnect: function(){
			// @todo: location.origin.replace(/^http/,"ws")+"/api..."
			this.sock = new WebSocket("ws://localhost:6001/api/v1/ws")
			this.sock.onopen = onopen
		},
		send: function(data) {
			if (this.sock === null || this.sock.readyState !== 1) {
				return new Promise((resolve, reject) => {
					reject("sock not ready")
				})
			}

			// assign an ID if not explicitly provided
			if (!("_id" in data)) {
				data._id = nextID++
			}
			data._id = ""+data._id


			// save the ID for calling back
			const id = data._id
			this.sock.send(JSON.stringify(data))
			return new Promise((resolve, reject) => {
				pending[id] = {resolve:resolve,reject:reject}
			})
		},
		connected: function() {
			return this.sock.readyState === 1
		},
		onready: function() {
			if (init_connect) {
				return new Promise((resolve, reject) => { resolve() })
			}
			return new Promise((resolve, reject) => {
				redies.push({resolve:resolve, reject:reject})
			})
		},
	}
	t.reconnect()

	inject('sock', t)
}

function reconnect() {

}