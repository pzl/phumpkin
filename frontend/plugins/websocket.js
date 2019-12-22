export default ({ app, store }, inject) => {
	const pending = {}
	let nextID = 1


	const onopen = event => {
		store.commit('socket/setConnected')
		console.log("socket opened")
		console.log(event)

		event.target.onmessage = onmessage
		event.target.onclose = onclose
		event.target.onerror = onerror
	}
	const onmessage = event => {
		console.log("got message")
		console.log(event)

		const j = JSON.parse(event.data)
		if ("_id" in j && j._id in pending) {
			pending[j._id](event.data)
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
	}

	const t = {
		sock: null,
		reconnect: function(){
			// @todo: location.origin.replace(/^http/,"ws")+"/api..."
			this.sock = new WebSocket("ws://localhost:6001/api/v1/ws")
			this.sock.onopen = onopen
		},
		send: function(data, cb) {
			if (this.sock === null || this.sock.readyState !== 1) {
				throw new Error('sock not ready')
			}

			// assign an ID if not explicitly provided
			if (!("_id" in data)) {
				data._id = nextID++
			}
			data._id = ""+data._id


			// save the ID for calling back
			const id = data._id
			this.sock.send(JSON.stringify(data))
			pending[id] = cb
		},
		connected: function() {
			return this.sock.readyState === 1
		},
	}
	t.reconnect()

	inject('sock', t)
}

function reconnect() {

}