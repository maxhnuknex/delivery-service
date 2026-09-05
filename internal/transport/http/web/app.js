const api = '/api/v1';
let cart = [];
let selectedRestaurant = null;
let activeFilter = '';
let currentUser = JSON.parse(localStorage.getItem('delivery-service-user') || 'null');

const rubles = (value) => `${new Intl.NumberFormat('ru-RU').format(value)} ₽`;
const byId = (id) => document.getElementById(id);
const open = (id) => { byId(id).classList.add('is-open'); byId(id).setAttribute('aria-hidden', 'false'); };
const close = (id) => { byId(id).classList.remove('is-open'); byId(id).setAttribute('aria-hidden', 'true'); };

async function request(path, options = {}) {
  const response = await fetch(`${api}${path}`, { headers: { 'Content-Type': 'application/json' }, ...options });
  if (!response.ok) throw new Error((await response.text()) || 'Не удалось выполнить запрос');
  return response.status === 204 ? null : response.json();
}

async function loadRestaurants() {
  const grid = byId('restaurant-grid');
  grid.innerHTML = '<div class="loading"><span></span>Загружаем заведения</div>';
  try {
    const query = activeFilter ? `?type=${activeFilter}` : '';
    const restaurants = await request(`/restaurants/${query}`);
    byId('restaurant-count').textContent = `${restaurants.length} ${plural(restaurants.length, 'заведение', 'заведения', 'заведений')}`;
    grid.innerHTML = restaurants.length ? restaurants.map((restaurant) => `
      <article class="restaurant-card">
        <div class="restaurant-art art-${restaurant.id % 3}"></div><span class="tag">${restaurant.type === 'SHOP' ? 'МАГАЗИН' : 'РЕСТОРАН'}</span>
        <h3>${escapeHtml(restaurant.name)}</h3><p>${escapeHtml(restaurant.address)}</p>
        <button type="button" data-restaurant-id="${restaurant.id}">Открыть меню →</button>
      </article>`).join('') : '<p class="muted">Пока нет активных заведений.</p>';
    grid.querySelectorAll('[data-restaurant-id]').forEach((button) => button.addEventListener('click', () => showMenu(restaurants.find((r) => r.id === Number(button.dataset.restaurantId)))));
  } catch (error) { grid.innerHTML = `<p class="muted">Не удалось загрузить каталог: ${escapeHtml(error.message)}</p>`; }
}

async function showMenu(restaurant) {
  selectedRestaurant = restaurant;
  const content = byId('menu-content');
  content.innerHTML = '<div class="loading"><span></span>Загружаем меню</div>';
  open('menu-modal');
  try {
    const menu = await request(`/restaurants/${restaurant.id}/menu`);
    content.innerHTML = `<div class="menu-head"><p class="eyebrow">${restaurant.type === 'SHOP' ? 'МАГАЗИН' : 'РЕСТОРАН'}</p><h2>${escapeHtml(restaurant.name)}</h2><p class="muted">${escapeHtml(restaurant.address)}</p></div><div class="menu-list">${menu.filter((item) => item.is_available).map((item) => `<article class="menu-item"><div><h3>${escapeHtml(item.name)}</h3><p>${escapeHtml(item.description || 'Готовится с заботой о деталях.')}</p></div><strong class="menu-price">${rubles(item.price)}</strong><button class="add-button" data-item-id="${item.id}">Добавить</button></article>`).join('') || '<p class="muted">В этом меню пока нет доступных позиций.</p>'}</div>`;
    content.querySelectorAll('[data-item-id]').forEach((button) => button.addEventListener('click', () => addToCart(menu.find((item) => item.id === Number(button.dataset.itemId)), restaurant)));
  } catch (error) { content.innerHTML = `<p class="muted">Не удалось загрузить меню: ${escapeHtml(error.message)}</p>`; }
}

function addToCart(item, restaurant) {
  if (cart.length && cart[0].restaurantId !== restaurant.id) { if (!confirm('Корзина относится к другому заведению. Очистить её и добавить этот товар?')) return; cart = []; }
  const existing = cart.find((entry) => entry.id === item.id);
  if (existing) existing.quantity += 1; else cart.push({ ...item, quantity: 1, restaurantId: restaurant.id, restaurantName: restaurant.name });
  renderCart();
}

function renderCart() {
  const total = cart.reduce((sum, item) => sum + item.price * item.quantity, 0);
  byId('cart-count').textContent = cart.reduce((sum, item) => sum + item.quantity, 0);
  byId('cart-total').textContent = rubles(total);
  byId('checkout-button').disabled = !cart.length;
  byId('cart-empty').hidden = Boolean(cart.length);
  byId('cart-items').innerHTML = cart.map((item) => `<article class="cart-item"><div><h3>${escapeHtml(item.name)}</h3><p>${rubles(item.price)} за шт.</p></div><strong>${rubles(item.price * item.quantity)}</strong><div class="quantity"><button data-cart-action="minus" data-id="${item.id}">−</button><span>${item.quantity}</span><button data-cart-action="plus" data-id="${item.id}">+</button></div></article>`).join('');
  byId('cart-items').querySelectorAll('[data-cart-action]').forEach((button) => button.addEventListener('click', () => changeQuantity(Number(button.dataset.id), button.dataset.cartAction === 'plus' ? 1 : -1)));
}

function changeQuantity(id, change) { const item = cart.find((entry) => entry.id === id); item.quantity += change; if (item.quantity < 1) cart = cart.filter((entry) => entry.id !== id); renderCart(); }
function plural(n, one, few, many) { const x = n % 100; const y = n % 10; return x > 10 && x < 20 ? many : y > 1 && y < 5 ? few : y === 1 ? one : many; }
function escapeHtml(text) { const node = document.createElement('div'); node.textContent = text; return node.innerHTML; }

byId('cart-button').addEventListener('click', () => open('cart-drawer'));
byId('checkout-button').addEventListener('click', () => { close('cart-drawer'); open('checkout-modal'); });
document.querySelectorAll('[data-filter]').forEach((button) => button.addEventListener('click', () => {
  activeFilter = button.dataset.filter;
  document.querySelectorAll('[data-filter]').forEach((item) => item.classList.toggle('active', item === button));
  loadRestaurants();
}));
document.querySelectorAll('[data-close]').forEach((button) => button.addEventListener('click', () => close(button.dataset.close)));
document.querySelectorAll('.modal,.drawer').forEach((element) => element.addEventListener('click', (event) => { if (event.target === element) close(element.id); }));

byId('checkout-form').addEventListener('submit', async (event) => {
  event.preventDefault();
  const button = event.target.querySelector('button'); button.disabled = true; button.textContent = 'Отправляем…';
  const data = new FormData(event.target);
  try {
    const user = await request('/users', { method: 'POST', body: JSON.stringify(Object.fromEntries(data)) });
    currentUser = user;
    localStorage.setItem('delivery-service-user', JSON.stringify(user));
    const order = await request('/orders/', { method: 'POST', body: JSON.stringify({ user_id: user.id, restaurant_id: cart[0].restaurantId, delivery_address: data.get('address'), items: cart.map((item) => ({ menu_item_id: item.id, quantity: item.quantity })) }) });
    cart = []; renderCart(); close('checkout-modal'); showOrder(order); event.target.reset();
  } catch (error) { alert(`Не удалось оформить заказ: ${error.message}`); } finally { button.disabled = false; button.textContent = 'Подтвердить заказ'; }
});

byId('orders-button').addEventListener('click', async () => {
  open('orders-modal');
  const content = byId('orders-content');
  if (!currentUser) { content.innerHTML = '<p class="muted">Оформите первый заказ — здесь появится история.</p>'; return; }
  content.innerHTML = '<div class="loading"><span></span>Загружаем заказы</div>';
  try {
    const orders = await request(`/users/${currentUser.id}/orders`);
    content.innerHTML = orders.length ? orders.map((order) => `<article class="order-row"><div><h3>Заказ #${order.id}</h3><p>${rubles(order.total_price)} · ${escapeHtml(order.delivery_address)}</p></div><span class="order-status">${escapeHtml(order.status)}</span></article>`).join('') : '<p class="muted">Заказов пока нет.</p>';
  } catch (error) { content.innerHTML = `<p class="muted">Не удалось загрузить историю: ${escapeHtml(error.message)}</p>`; }
});

function showOrder(order) {
  const render = (current) => {
    const stages = [['CREATED','Принят'],['ACCEPTED','В ресторане'],['IN_PROGRESS','Готовится'],['READY_FOR_DELIVERY','В пути'],['DELIVERED','Доставлен']];
    const index = Math.max(0, stages.findIndex(([status]) => status === current.status));
    byId('order-content').innerHTML = `<div class="order-mark">✓</div><p class="eyebrow">ЗАКАЗ #${current.id}</p><h2>Заказ оформлен</h2><p class="muted">Статус: ${stages[index][1] || current.status}<br />На сумму ${rubles(current.total_price)}</p><div class="status-line">${stages.map(([status,label], i) => `<span class="status-step ${i <= index ? 'active' : ''}">${label}</span>`).join('')}</div>`;
  };
  render(order); open('order-modal');
  const timer = setInterval(async () => { try { const updated = await request(`/orders/${order.id}`); render(updated); if (['DELIVERED','CANCELLED'].includes(updated.status)) clearInterval(timer); } catch (_) { clearInterval(timer); } }, 4000);
}

loadRestaurants();
