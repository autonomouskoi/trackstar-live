import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";

import * as livepb from "/m/trackstar-live/pb/trackstar-live/live_pb.js";
import * as tspb from "/m/trackstar/pb/trackstar_pb.js";
import { ControlPanel } from "/tk.js";

const TOPIC_EVENT = enumName(livepb.BusTopic, livepb.BusTopic.TRACKSTAR_LIVE_EVENT);

let help = document.createElement('div');
help.innerHTML = `
<p>
Shows the last track TS Live sent, or the last error that occurred sending a track.
</p>
`;

class Status extends ControlPanel {
    private _artistDiv: HTMLDivElement;
    private _titleDiv: HTMLDivElement;
    private _errorDiv: HTMLDivElement;

    constructor() {
        super({ title: 'Status', help });

        this.innerHTML = `
<style>
label {
    font-weight: bold;
}
</style>
<div class="grid-2-col">
    <label for="artist">Last Track Artist</label>
    <div id="artist"></div>

    <label for="title">Last Track Title</label>
    <div id="title"></div>

    <label for="error">Last Error</label>
    <div id="error"></div>
</div>
`;

        this._artistDiv = this.querySelector('div#artist');
        this._titleDiv = this.querySelector('div#title');
        this._errorDiv = this.querySelector('div#error');

        bus.subscribe(TOPIC_EVENT, (msg) => this._handleTSEvent(msg));
    }

    private _handleTSEvent(msg: buspb.BusMessage) {
        if (msg.type !== livepb.MessageTypeEvent.TRACK_SEND_EVENT) {
            return;
        }
        if (msg.error) {
            this._errorDiv.innerText = msg.error.userMessage;
            return;
        }
        let tu = tspb.TrackUpdate.fromBinary(msg.message);
        this._artistDiv.innerText = tu.track.artist;
        this._titleDiv.innerText = tu.track.title;
    }
}
customElements.define('trackstar-live-status', Status, { extends: 'fieldset' });

export { Status };