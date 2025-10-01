class SetsList extends HTMLOListElement {
    constructor(getSets: Promise<number[]>, onClick: (setID: number) => void) {
        super();

        getSets.then((sets) => {
            sets.forEach((setID) => {
                let a = document.createElement('a');
                a.href = '#';
                a.innerText = new Date(setID).toLocaleString();
                a.addEventListener('click', () => onClick(setID));
                let li = document.createElement('li');
                li.id = setID.toString();
                li.appendChild(a);
                this.appendChild(li);
            });
        })
    }

    selectSet(setID: number) {
        let id = setID.toString();
        let children = this.children;
        for (let i = 0; i < children.length; i++) {
            let li = children.item(i);
            if (li.id === id) {
                li.classList.add('selected');
            } else {
                li.classList.remove('selected');
            }
        }
    }
}
customElements.define('tslive-setslist', SetsList, { extends: 'ol' });

export { SetsList };