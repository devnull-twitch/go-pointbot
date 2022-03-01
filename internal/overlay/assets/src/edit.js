const assignEditEvents = (box) => {
    const inputs = box.querySelectorAll('.js-edit-input');
    for (let i = 0; i < inputs.length; i++)
    {
        inputs.item(i).addEventListener('change', () => {
            switch (inputs.item(i).id) {
                case 'left':
                    box.dataset.left = inputs.item(i).value;
                    box.dataset.right = "0";
                    box.querySelector('#right').value = 0;
                    break;
                case 'top':
                    box.dataset.top = inputs.item(i).value;
                    box.dataset.bottom = "0";
                    box.querySelector('#bottom').value = 0;
                    break;
                case 'right':
                    box.dataset.right = inputs.item(i).value;
                    box.dataset.left = "0";
                    box.querySelector('#left').value = 0;
                    break;
                case 'bottom':
                    box.dataset.bottom = inputs.item(i).value;
                    box.dataset.top = "0";
                    box.querySelector('#top').value = 0;
                    break;
                case 'bgcolor':
                    box.dataset.bgcolor = inputs.item(i).value;
                    break;
            }
        });
    }
};
const main = () => {
    const boxes = document.querySelectorAll('.js-pb-box');
    for (let i = 0; i < boxes.length; i++) {
        const box = boxes.item(i);
        const editbox = box.querySelector('.js-edit-box');
        assignEditEvents(box);
        editbox.addEventListener('click', (e) => {
            e.stopPropagation();
        });
        editbox.addEventListener('dblclick', (e) => {
            e.stopPropagation();
        });
        box.addEventListener('click', () => {
            if (editbox.classList.contains('hidden')) {
                editbox.classList.remove('hidden');
            } else {
                editbox.classList.add('hidden');
            }
        });
        box.addEventListener('dblclick', () => {
            if (editbox.classList.contains('translate-x-full')) {
                editbox.classList.remove('translate-x-full');
                editbox.classList.add('-translate-x-full');
                editbox.classList.remove('right-0');
            } else {
                editbox.classList.remove('-translate-x-full');
                editbox.classList.add('translate-x-full');
                editbox.classList.add('right-0');
            }
        });
    }

    const newBox = document.querySelector('.js-new-box');
    document.addEventListener('keydown', (e) => {
        if (e.altKey && e.key === 'n') {
            newBox.classList.toggle('hidden');
        }
    })
};
main();