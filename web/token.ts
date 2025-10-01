import { bus, enumName } from "/bus.js";
import * as buspb from "/pb/bus/bus_pb.js";
import * as livepb from "/m/trackstar-live/pb/trackstar-live/live_pb.js";
import { Cfg } from './controller.js';
import { UpdatingControlPanel } from '/tk.js';

const TOPIC_COMMAND = enumName(livepb.BusTopic, livepb.BusTopic.TRACKSTAR_LIVE_COMMAND);

let help = document.createElement('div');
help.innerHTML = `
<p>
For Trackstar Live! to send track updates it needs a <em>token</em>. This is a big piece of text
provided by whoever is operating the Trackstar Live! website you're using. The token specifies
your user ID, the site to talk to, and the credetials to authenticate you. Never share the value
given to you by the site operator with someone else!
</p>
`;

interface setTokenParams {
    label?: string,
    rawToken?: string,
    enabled?: boolean
}

class Tokens extends UpdatingControlPanel<livepb.Config> {
    private _tokensDiv: HTMLDivElement;
    private _setTokenDialog: SetTokenDialog;

    private _cfg: Cfg;

    constructor(cfg: Cfg) {
        super({ title: 'Tokens', help, data: cfg });
        this._cfg = cfg;

        this.innerHTML = `
<div class="flex-column" style="gap: 1rem" id="tokens"></div>
`;

        this._tokensDiv = this.querySelector('div#tokens');
        this._setTokenDialog = new SetTokenDialog((rawToken) => this._setToken({ rawToken }));
        this._setTokenDialog.close();
        this.appendChild(this._setTokenDialog);

        let newButton = document.createElement('button');
        newButton.innerText = '+';
        newButton.addEventListener('click', () => this._setTokenDialog.showModal());
        newButton.style.width = '5rem';
        this.appendChild(newButton);
    }

    update(cfg: livepb.Config) {
        this._tokensDiv.textContent = '';
        Object.keys(cfg.tokens).toSorted().forEach((label) => {
            this._tokensDiv.appendChild(new Token({
                label,
                tokenCfg: cfg.tokens[label],
                onEnabled: (enabled) => this._setToken({ label, enabled }),
                onDelete: () => this._setToken({ label }),
            }));
        });
    }

    private _setToken(params: setTokenParams) {
        let msg = new buspb.BusMessage({
            topic: TOPIC_COMMAND,
            type: livepb.MessageTypeCommand.TOKEN_SET_REQ,
            message: new livepb.TokenSetRequest(params).toBinary(),
        });
        bus.sendAnd(msg).then((reply) => {
            this._setTokenDialog.close();
            this._cfg.refresh();
        })
    }
}
customElements.define('trackstar-live-tokens', Tokens, { extends: 'fieldset' });

class Token extends HTMLDetailsElement {

    constructor({ label, tokenCfg, onEnabled, onDelete }:
        {
            label: string,
            tokenCfg: livepb.TokenConfig,
            onEnabled: (en: boolean) => void,
            onDelete: () => void,
        }
    ) {
        super();

        this.innerHTML = `
<summary>
        <label for="enabled">Enabled:</label>
        <input id="enabled" type="checkbox" />
        <a href=""></a>
</summary>
<div class="grid-2-col" id="details"></div>
`

        let token = tokenCfg.token;
        if (!token) {
            return
        }

        let details = this.querySelector('div#details');

        let add = (id: string, label: string, content: string): HTMLInputElement => {
            let l = document.createElement('label');
            l.htmlFor = id;
            l.innerText = label;
            details.appendChild(l);

            let i = document.createElement('input');
            i.id = id;
            i.disabled = true;
            i.value = content;
            details.appendChild(i);

            return i;
        };

        add('user_id', 'User ID', token.subject);
        add('issuer', 'Issuer', token.issuer);
        add('audience', 'Audience', `${token.audience}`);
        add('issued', 'Issued', new Date(Number(token.issuedAt)).toLocaleString());
        add('expires', 'Expires', new Date(Number(token.expiresAt)).toLocaleString());

        let enabled: HTMLInputElement = this.querySelector('summary>input');
        enabled.checked = tokenCfg.enabled;
        enabled.addEventListener('change', () => onEnabled(enabled.checked));
        let a: HTMLAnchorElement = this.querySelector('summary>a');
        a.innerText = label;
        a.href = label;

        let deleteButton = document.createElement('button');
        deleteButton.type = 'button';
        deleteButton.innerText = 'Delete';
        deleteButton.addEventListener('click', () => {
            if (confirm(`Delete ${label}?`)) {
                onDelete();
            }
        });
        details.appendChild(deleteButton);
    }
}
customElements.define('trackstar-live-token', Token, { extends: 'details' });

class SetTokenDialog extends HTMLDialogElement {
    private _set = (token: string) => { };
    private _ta: HTMLTextAreaElement;

    constructor(set: (token: string) => void) {
        super();
        this._set = set;

        this.innerHTML = `
<div class="flex-column">
    <h2>Set Token</h2>
    <textarea cols="40" rows="8"></textarea>
    <button id="save" type="button">Save</button>
    <button id="cancel" type="button">Cancel</button>
</div>`;

        this._ta = this.querySelector('textarea');

        let cancel: HTMLButtonElement = this.querySelector('button#cancel');
        cancel.addEventListener('click', () => this._close());

        let save: HTMLButtonElement = this.querySelector('button#save');
        save.addEventListener('click', () => this._save());
    }

    private _save() {
        this._set(this._ta.value);
        this._close();
    }

    private _close() {
        this._ta.value = '';
        this.close();
    }
}
customElements.define('trackstar-live-token-set', SetTokenDialog, { extends: 'dialog' });

export { Tokens };