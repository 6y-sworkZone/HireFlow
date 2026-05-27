const API_BASE = '/api';
let token = localStorage.getItem('token') || '';
let currentUser = null;
let currentPage = 'dashboard';

const pageTitles = {
  dashboard: '仪表盘', jobs: '职位管理', candidates: '候选人管理',
  workflow: '招聘流程', interviews: '面试安排', offers: 'Offer管理',
  analytics: '数据分析', settings: '系统设置'
};

function api(url, options = {}) {
  const headers = { 'Content-Type': 'application/json', ...options.headers };
  if (token) headers['Authorization'] = 'Bearer ' + token;
  return fetch(API_BASE + url, { ...options, headers }).then(r => r.json());
}

function showToast(msg, type = 'success') {
  const t = document.getElementById('toast');
  t.textContent = msg; t.className = 'toast ' + type;
  t.style.display = 'block';
  setTimeout(() => { t.style.display = 'none'; }, 3000);
}

function showModal(html) {
  document.getElementById('modalContent').innerHTML = html;
  document.getElementById('modal').style.display = 'flex';
}

function closeModal() { document.getElementById('modal').style.display = 'none'; }

function statusText(s) {
  return { draft:'草稿',open:'开放',paused:'暂停',closed:'关闭',pool:'候选人池',active:'进行中',rejected:'已拒绝',scheduled:'待进行',completed:'已完成',pending_approval:'待审批',approved:'已审批',sent:'已发送',accepted:'已接受',rejected_offer:'已拒绝',expired:'已过期' }[s] || s;
}

function fmt(s) { return s ? s.substring(0, 16).replace('T', ' ') : '-'; }

function showLoginPage() {
  document.body.innerHTML = `<div class="login-page"><div class="login-card">
    <div class="login-logo"><div class="logo-icon" style="margin:0 auto;width:50px;height:50px;font-size:20px;">HF</div><h2>HireFlow</h2><p style="color:#888;font-size:13px;">智能招聘管理系统</p></div>
    <form onsubmit="handleLogin(event)">
      <div class="form-group"><label class="form-label">用户名</label><input class="form-input" id="lu" value="admin"></div>
      <div class="form-group"><label class="form-label">密码</label><input type="password" class="form-input" id="lp" value="admin123"></div>
      <button type="submit" class="btn btn-primary" style="width:100%;margin-top:10px;">登录</button>
    </form>
    <p style="text-align:center;margin-top:15px;font-size:12px;color:#888;">还没有账号？<a href="#" onclick="showRegisterPage()" style="color:#667eea;">立即注册</a></p>
  </div></div>`;
}

function handleLogin(e) {
  e.preventDefault();
  api('/auth/login', { method: 'POST', body: JSON.stringify({ username: document.getElementById('lu').value, password: document.getElementById('lp').value }) }).then(r => {
    if (r.code === 0) { token = r.data.token; currentUser = r.data.user; localStorage.setItem('token', token); initApp(); }
    else showToast(r.message, 'error');
  });
}

function showRegisterPage() {
  showModal(`<div class="modal-header"><h3 class="modal-title">注册账号</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
    <form onsubmit="handleRegister(event)">
      <div class="form-group"><label class="form-label">用户名 *</label><input class="form-input" id="ru" required></div>
      <div class="form-group"><label class="form-label">邮箱 *</label><input type="email" class="form-input" id="re" required></div>
      <div class="form-group"><label class="form-label">密码 * (至少6位)</label><input type="password" class="form-input" id="rp" required></div>
      <div class="form-group"><label class="form-label">真实姓名 *</label><input class="form-input" id="rn" required></div>
      <div class="form-row">
        <div class="form-group"><label class="form-label">部门</label><input class="form-input" id="rd"></div>
        <div class="form-group"><label class="form-label">角色</label>
          <select class="form-select" id="rr"><option value="HR">HR</option><option value="面试官">面试官</option><option value="部门经理">部门经理</option><option value="admin">管理员</option></select>
        </div>
      </div>
      <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">注册</button></div>
    </form>`);
}

function handleRegister(e) {
  e.preventDefault();
  api('/auth/register', { method: 'POST', body: JSON.stringify({ username: document.getElementById('ru').value, email: document.getElementById('re').value, password: document.getElementById('rp').value, real_name: document.getElementById('rn').value, department: document.getElementById('rd').value, role: document.getElementById('rr').value }) }).then(r => {
    if (r.code === 0) { showToast('注册成功，请登录'); closeModal(); } else showToast(r.message, 'error');
  });
}

function logout() { localStorage.removeItem('token'); token = ''; currentUser = null; showLoginPage(); }

function initApp() {
  api('/auth/me').then(r => {
    if (r.code === 0) { currentUser = r.data; renderLayout(); updateUser(); setupNav(); loadUnread(); navigate(window.location.pathname.slice(1) || 'dashboard'); }
    else showLoginPage();
  });
}

function renderLayout() {
  document.body.innerHTML = `<div class="layout">
    <aside class="sidebar"><div class="logo"><span class="logo-icon">HF</span><span class="logo-text">HireFlow</span></div>
      <nav class="nav-menu">
        <a href="/dashboard" class="nav-item" data-page="dashboard"><span class="nav-icon">📊</span> 仪表盘</a>
        <a href="/jobs" class="nav-item" data-page="jobs"><span class="nav-icon">💼</span> 职位管理</a>
        <a href="/candidates" class="nav-item" data-page="candidates"><span class="nav-icon">👥</span> 候选人管理</a>
        <a href="/workflow" class="nav-item" data-page="workflow"><span class="nav-icon">🔄</span> 招聘流程</a>
        <a href="/interviews" class="nav-item" data-page="interviews"><span class="nav-icon">📅</span> 面试安排</a>
        <a href="/offers" class="nav-item" data-page="offers"><span class="nav-icon">📄</span> Offer管理</a>
        <a href="/analytics" class="nav-item" data-page="analytics"><span class="nav-icon">📈</span> 数据分析</a>
        <a href="/settings" class="nav-item" data-page="settings"><span class="nav-icon">⚙️</span> 系统设置</a>
      </nav>
    </aside>
    <main class="main">
      <header class="header">
        <div class="header-left"><h1 class="page-title" id="pt">仪表盘</h1></div>
        <div class="header-right">
          <div class="notification-bell" onclick="toggleNotif()">🔔<span class="badge" id="uc" style="display:none">0</span></div>
          <div class="user-info">
            <div class="avatar-circle" id="ua">A</div>
            <div class="user-detail"><span class="user-name" id="un">管理员</span><span class="user-role" id="ur">admin</span></div>
            <button class="btn-logout" onclick="logout()">退出</button>
          </div>
        </div>
      </header>
      <div class="content" id="ct"></div>
    </main>
  </div>
  <div id="nd" class="notification-dropdown" style="display:none">
    <div class="notification-header"><span>通知</span><button onclick="readAll()" class="link-btn">全部已读</button></div>
    <div id="nl" class="notification-list"></div>
  </div>
  <div id="modal" class="modal" style="display:none"><div class="modal-overlay" onclick="closeModal()"></div><div class="modal-content" id="mc"></div></div>
  <div id="toast" class="toast" style="display:none"></div>`;
}

function updateUser() {
  if (currentUser) {
    document.getElementById('un').textContent = currentUser.real_name || currentUser.username;
    document.getElementById('ur').textContent = currentUser.role;
    document.getElementById('ua').textContent = (currentUser.real_name || currentUser.username).charAt(0);
  }
}

function setupNav() {
  document.querySelectorAll('.nav-item').forEach(i => {
    i.addEventListener('click', e => { e.preventDefault(); navigate(i.dataset.page); });
  });
}

function navigate(page) {
  currentPage = page;
  document.querySelectorAll('.nav-item').forEach(i => i.classList.toggle('active', i.dataset.page === page));
  document.getElementById('pt').textContent = pageTitles[page] || page;
  history.pushState({}, '', '/' + page);
  document.getElementById('ct').innerHTML = '<div class="card"><p>加载中...</p></div>';
  const fns = { dashboard: renderDashboard, jobs: renderJobs, candidates: renderCandidates, workflow: renderWorkflow, interviews: renderInterviews, offers: renderOffers, analytics: renderAnalytics, settings: renderSettings };
  (fns[page] || renderDashboard)();
}

function toggleNotif() {
  const d = document.getElementById('nd');
  d.style.display = d.style.display === 'none' ? 'block' : 'none';
  if (d.style.display === 'block') loadNotif();
}

function loadNotif() {
  api('/notifications').then(r => {
    if (r.code === 0 && r.data) {
      document.getElementById('nl').innerHTML = r.data.length === 0
        ? '<div style="padding:20px;text-align:center;color:#999;">暂无通知</div>'
        : r.data.map(n => `<div class="notification-item ${n.is_read ? '' : 'unread'}" onclick="readNotif(${n.id})"><div class="notification-title">${n.title}</div><div class="notification-content">${n.content || ''}</div><div style="font-size:11px;color:#aaa;margin-top:4px;">${fmt(n.created_at)}</div></div>`).join('');
    }
  });
}

function readNotif(id) { api('/notifications/' + id + '/read', { method: 'PUT' }).then(() => { loadNotif(); loadUnread(); }); }
function readAll() { api('/notifications/read-all', { method: 'PUT' }).then(() => { loadNotif(); loadUnread(); }); }

function loadUnread() {
  api('/notifications/unread-count').then(r => {
    if (r.code === 0 && r.data) {
      const b = document.getElementById('uc');
      if (r.data.unread_count > 0) { b.textContent = r.data.unread_count; b.style.display = 'inline'; } else b.style.display = 'none';
    }
  });
}

function renderDashboard() {
  Promise.all([api('/analytics/monthly-trend'), api('/analytics/offer-acceptance'), api('/analytics/department-progress')]).then(([t, o, d]) => {
    const td = t.data || []; const l = td[td.length - 1] || {};
    document.getElementById('ct').innerHTML = `<div class="stat-cards">
      <div class="stat-card"><div class="label">本月新增简历</div><div class="value">${l.new_candidates || 0}</div></div>
      <div class="stat-card"><div class="label">本月面试安排</div><div class="value">${l.interviews || 0}</div></div>
      <div class="stat-card"><div class="label">本月Offer发放</div><div class="value">${l.offers || 0}</div></div>
      <div class="stat-card"><div class="label">本月成功入职</div><div class="value">${l.hired || 0}</div></div>
      <div class="stat-card"><div class="label">Offer接受率</div><div class="value">${(o.data?.acceptance_rate || 0).toFixed(1)}%</div></div>
    </div>
    <div class="card"><div class="card-header"><h3 class="card-title">月度招聘趋势</h3></div>
      <table><thead><tr><th>月份</th><th>新增简历</th><th>面试数量</th><th>Offer发放</th><th>成功入职</th></tr></thead><tbody>
        ${td.map(m => `<tr><td>${m.month}</td><td>${m.new_candidates}</td><td>${m.interviews}</td><td>${m.offers}</td><td>${m.hired}</td></tr>`).join('')}
      </tbody></table>
    </div>
    <div class="card"><div class="card-header"><h3 class="card-title">部门招聘进度</h3></div>
      <table><thead><tr><th>部门</th><th>职位数</th><th>候选人数</th><th>Offer数</th><th>入职数</th></tr></thead><tbody>
        ${(d.data || []).map(x => `<tr><td>${x.department || '-'}</td><td>${x.total_jobs}</td><td>${x.total_candidates}</td><td>${x.total_offers}</td><td>${x.total_hired}</td></tr>`).join('')}
      </tbody></table>
    </div>`;
  });
}

let jPage = 1, jFilter = {};

function renderJobs() {
  document.getElementById('ct').innerHTML = `<div class="card">
    <div class="card-header"><h3 class="card-title">职位列表</h3>
      <div class="btn-group"><button class="btn btn-secondary" onclick="showTemplates()">职位模板</button><button class="btn btn-primary" onclick="showJobForm()">+ 发布职位</button></div>
    </div>
    <div class="filter-bar">
      <div class="form-group"><input class="form-input" placeholder="搜索" id="jk"></div>
      <div class="form-group"><select class="form-select" id="jd"><option value="">全部部门</option></select></div>
      <div class="form-group"><select class="form-select" id="js"><option value="">全部状态</option><option value="draft">草稿</option><option value="open">开放</option><option value="paused">暂停</option><option value="closed">关闭</option></select></div>
      <button class="btn btn-secondary" onclick="applyJF()">筛选</button>
    </div>
    <div id="jl">加载中...</div>
  </div>`;
  api('/jobs/departments').then(r => { if (r.code === 0 && r.data) r.data.forEach(d => { document.getElementById('jd').innerHTML += `<option value="${d}">${d}</option>`; }); });
  loadJobs();
}

function applyJF() { jFilter = { keyword: document.getElementById('jk').value, department: document.getElementById('jd').value, status: document.getElementById('js').value }; jPage = 1; loadJobs(); }

function loadJobs() {
  let url = `/jobs?page=${jPage}&page_size=10`;
  if (jFilter.keyword) url += `&keyword=${encodeURIComponent(jFilter.keyword)}`;
  if (jFilter.department) url += `&department=${jFilter.department}`;
  if (jFilter.status) url += `&status=${jFilter.status}`;
  api(url).then(r => {
    if (r.code === 0 && r.data) {
      const { list, pagination } = r.data;
      document.getElementById('jl').innerHTML = list.length === 0
        ? '<div class="empty-state"><div class="empty-state-icon">📭</div>暂无职位</div>'
        : `<table><thead><tr><th>职位名称</th><th>部门</th><th>薪资范围</th><th>标签</th><th>状态</th><th>发布人</th><th>发布时间</th><th>操作</th></tr></thead><tbody>
          ${list.map(j => `<tr><td><strong>${j.title}</strong></td><td>${j.department}</td><td>${j.salary_min}K - ${j.salary_max}K</td><td>${(j.tags||'').split(',').filter(t=>t).map(t=>`<span class="tag">${t}</span>`).join('')}</td><td><span class="status-badge status-${j.status}">${statusText(j.status)}</span></td><td>${j.creator_name||'-'}</td><td>${fmt(j.created_at)}</td><td><button class="btn btn-sm btn-secondary" onclick="viewJob(${j.id})">查看</button> <button class="btn btn-sm btn-secondary" onclick="showJobForm(${j.id})">编辑</button> <button class="btn btn-sm btn-danger" onclick="delJob(${j.id})">删除</button></td></tr>`).join('')}
        </tbody></table>
        <div class="pagination">${pagination.page > 1 ? `<button onclick="jPage--;loadJobs()">上一页</button>` : ''}<button class="active">${pagination.page}</button>${pagination.page * pagination.page_size < pagination.total ? `<button onclick="jPage++;loadJobs()">下一页</button>` : ''}<span>共 ${pagination.total} 条</span></div>`;
    }
  });
}

function showJobForm(id) { if (id) api('/jobs/' + id).then(r => { if (r.code === 0) showJobFormModal(r.data); }); else showJobFormModal(null); }

function showJobFormModal(j) {
  showModal(`<div class="modal-header"><h3 class="modal-title">${j ? '编辑职位' : '发布职位'}</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
    <form onsubmit="saveJob(event, ${j?.id||'null'})">
      <div class="form-group"><label class="form-label">职位名称 *</label><input class="form-input" id="jft" value="${j?.title||''}" required></div>
      <div class="form-row">
        <div class="form-group"><label class="form-label">部门 *</label><input class="form-input" id="jfd" value="${j?.department||''}" required></div>
        <div class="form-group"><label class="form-label">工作地点</label><input class="form-input" id="jfl" value="${j?.location||''}"></div>
      </div>
      <div class="form-row">
        <div class="form-group"><label class="form-label">薪资下限(K)</label><input type="number" class="form-input" id="jfsm" value="${j?.salary_min||''}"></div>
        <div class="form-group"><label class="form-label">薪资上限(K)</label><input type="number" class="form-input" id="jfxm" value="${j?.salary_max||''}"></div>
      </div>
      <div class="form-group"><label class="form-label">职位描述(Markdown) - 支持实时预览</label>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
          <textarea class="form-textarea" id="jfdesc" rows="8" oninput="updateMDPreview()" style="height:200px;">${j?.description||''}</textarea>
          <div class="md-preview" id="mdprev" style="border:1px solid #ddd;border-radius:6px;padding:12px;height:200px;overflow-y:auto;background:#fafafa;"></div>
        </div>
      </div>
      <div class="form-group"><label class="form-label">任职要求</label><textarea class="form-textarea" id="jfreq" rows="4">${j?.requirements||''}</textarea></div>
      <div class="form-row">
        <div class="form-group"><label class="form-label">标签(逗号分隔)</label><input class="form-input" id="jftg" value="${j?.tags||''}"></div>
        <div class="form-group"><label class="form-label">状态</label>
          <select class="form-select" id="jfst"><option value="draft" ${j?.status==='draft'?'selected':''}>草稿</option><option value="open" ${j?.status==='open'?'selected':''}>开放</option><option value="paused" ${j?.status==='paused'?'selected':''}>暂停</option><option value="closed" ${j?.status==='closed'?'selected':''}>关闭</option></select>
        </div>
      </div>
      <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">保存</button></div>
    </form>`);
  setTimeout(updateMDPreview, 50);
}

function updateMDPreview() {
  const md = document.getElementById('jfdesc')?.value || '';
  const prev = document.getElementById('mdprev');
  if (prev && typeof marked !== 'undefined') {
    prev.innerHTML = marked.parse(md) || '<p style="color:#999;">预览区域</p>';
  }
}

function saveJob(e, id) {
  e.preventDefault();
  const d = { title: document.getElementById('jft').value, department: document.getElementById('jfd').value, location: document.getElementById('jfl').value, salary_min: parseInt(document.getElementById('jfsm').value)||0, salary_max: parseInt(document.getElementById('jfxm').value)||0, description: document.getElementById('jfdesc').value, requirements: document.getElementById('jfreq').value, tags: document.getElementById('jftg').value, status: document.getElementById('jfst').value };
  api(id ? '/jobs/' + id : '/jobs', { method: id ? 'PUT' : 'POST', body: JSON.stringify(d) }).then(r => {
    if (r.code === 0) { showToast('保存成功'); closeModal(); loadJobs(); } else showToast(r.message, 'error');
  });
}

function viewJob(id) {
  api('/jobs/' + id).then(r => {
    if (r.code === 0) {
      const j = r.data;
      api('/jobs/' + id + '/stats').then(sr => {
        const s = sr.code === 0 ? sr.data : {};
        const descHTML = (typeof marked !== 'undefined' && j.description) ? marked.parse(j.description) : (j.description || '暂无描述');
        const reqHTML = (typeof marked !== 'undefined' && j.requirements) ? marked.parse(j.requirements) : (j.requirements || '暂无要求');
        showModal(`<div class="modal-header"><h3 class="modal-title">${j.title}</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
          <div style="line-height:1.8;"><p><strong>部门：</strong>${j.department}</p><p><strong>工作地点：</strong>${j.location||'-'}</p><p><strong>薪资范围：</strong>${j.salary_min}K - ${j.salary_max}K</p><p><strong>标签：</strong>${(j.tags||'').split(',').filter(t=>t).map(t=>`<span class="tag">${t}</span>`).join('')}</p><p><strong>状态：</strong><span class="status-badge status-${j.status}">${statusText(j.status)}</span></p><p><strong>发布人：</strong>${j.creator_name||'-'}</p><p><strong>发布时间：</strong>${fmt(j.created_at)}</p><hr style="margin:15px 0;"><h4>职位描述</h4><div class="md-preview">${descHTML}</div><h4>任职要求</h4><div class="md-preview">${reqHTML}</div><hr style="margin:15px 0;"><h4>招聘统计</h4><p>收到简历：${s.resume_count||0} 份</p><p>面试数量：${s.interview_count||0} 场</p><p>Offer数量：${s.offer_count||0} 个</p></div>
          <div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`);
      });
    }
  });
}

function delJob(id) { if (!confirm('确定删除？')) return; api('/jobs/' + id, { method: 'DELETE' }).then(r => { if (r.code === 0) { showToast('删除成功'); loadJobs(); } }); }

function showTemplates() {
  api('/jobs/templates').then(r => {
    const ts = r.code === 0 ? (r.data || []) : [];
    showModal(`<div class="modal-header"><h3 class="modal-title">职位模板</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
      <div style="margin-bottom:15px;"><button class="btn btn-primary btn-sm" onclick="showTplForm()">+ 新建模板</button></div>
      ${ts.length===0 ? '<div class="empty-state">暂无模板</div>' : ts.map(t => `<div style="padding:12px;border:1px solid #eee;border-radius:6px;margin-bottom:10px;"><div style="display:flex;justify-content:space-between;align-items:center;"><div><strong>${t.name}</strong><span style="color:#888;margin-left:10px;">${t.title} / ${t.department}</span></div><div><button class="btn btn-sm btn-secondary" onclick="useTpl(${t.id})">使用</button> <button class="btn btn-sm btn-danger" onclick="delTpl(${t.id})">删除</button></div></div></div>`).join('')}
      <div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`);
  });
}

function showTplForm() {
  showModal(`<div class="modal-header"><h3 class="modal-title">新建模板</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
    <form onsubmit="saveTpl(event)">
      <div class="form-group"><label class="form-label">模板名称 *</label><input class="form-input" id="tn" required></div>
      <div class="form-row"><div class="form-group"><label class="form-label">职位名称 *</label><input class="form-input" id="tt" required></div><div class="form-group"><label class="form-label">部门 *</label><input class="form-input" id="td" required></div></div>
      <div class="form-group"><label class="form-label">职位描述</label><textarea class="form-textarea" id="tdesc" rows="3"></textarea></div>
      <div class="form-group"><label class="form-label">任职要求</label><textarea class="form-textarea" id="treq" rows="3"></textarea></div>
      <div class="form-group"><label class="form-label">标签(逗号分隔)</label><input class="form-input" id="ttg"></div>
      <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">保存</button></div>
    </form>`);
}

function saveTpl(e) {
  e.preventDefault();
  api('/jobs/templates', { method: 'POST', body: JSON.stringify({ name: document.getElementById('tn').value, title: document.getElementById('tt').value, department: document.getElementById('td').value, description: document.getElementById('tdesc').value, requirements: document.getElementById('treq').value, tags: document.getElementById('ttg').value }) }).then(r => {
    if (r.code === 0) { showToast('模板创建成功'); closeModal(); }
  });
}

function useTpl(id) {
  api('/jobs/templates').then(r => {
    if (r.code === 0) { const t = r.data.find(x => x.id === id); if (t) { closeModal(); showJobFormModal({ title: t.title, department: t.department, location: t.location, salary_min: t.salary_min, salary_max: t.salary_max, description: t.description, requirements: t.requirements, tags: t.tags, status: 'draft' }); } }
  });
}

function delTpl(id) { if (!confirm('确定删除？')) return; api('/jobs/templates/' + id, { method: 'DELETE' }).then(r => { if (r.code === 0) { showToast('删除成功'); showTemplates(); } }); }

function renderCandidates() {
  document.getElementById('ct').innerHTML = `<div class="card">
    <div class="card-header"><h3 class="card-title">候选人管理</h3>
      <div class="btn-group"><button class="btn btn-secondary" onclick="renderPool()">候选人池</button><button class="btn btn-primary" onclick="showCandForm()">+ 录入候选人</button></div>
    </div>
    <div class="filter-bar">
      <div class="form-group"><input class="form-input" placeholder="搜索姓名/邮箱/电话" id="ck"></div>
      <div class="form-group"><select class="form-select" id="cs"><option value="">全部来源</option><option value="招聘网站">招聘网站</option><option value="内推">内推</option><option value="猎头">猎头</option><option value="主动投递">主动投递</option></select></div>
      <button class="btn btn-secondary" onclick="applyCF()">筛选</button>
    </div>
    <div id="cl">加载中...</div>
  </div>`;
  loadCands();
}

function applyCF() {
  let url = '/candidates?page=1&page_size=20';
  const k = document.getElementById('ck').value, s = document.getElementById('cs').value;
  if (k) url += `&keyword=${encodeURIComponent(k)}`;
  if (s) url += `&source=${s}`;
  loadCands(url);
}

function loadCands(u = '/candidates?page=1&page_size=20') {
  api(u).then(r => {
    if (r.code === 0 && r.data) {
      const list = r.data.list;
      document.getElementById('cl').innerHTML = list.length === 0
        ? '<div class="empty-state"><div class="empty-state-icon">👥</div>暂无候选人</div>'
        : `<table><thead><tr><th>姓名</th><th>邮箱</th><th>电话</th><th>当前公司</th><th>工作年限</th><th>期望薪资</th><th>来源</th><th>状态</th><th>操作</th></tr></thead><tbody>
          ${list.map(c => `<tr><td><strong>${c.name}</strong></td><td>${c.email||'-'}</td><td>${c.phone||'-'}</td><td>${c.current_company||'-'}</td><td>${c.work_years}年</td><td>${c.expected_salary}K</td><td><span class="tag">${c.source}</span></td><td><span class="status-badge status-${c.status}">${statusText(c.status)}</span></td><td><button class="btn btn-sm btn-secondary" onclick="viewCand(${c.id})">详情</button> <button class="btn btn-sm btn-secondary" onclick="showCandForm(${c.id})">编辑</button> <button class="btn btn-sm btn-secondary" onclick="assignJob(${c.id})">关联职位</button></td></tr>`).join('')}
        </tbody></table>`;
    }
  });
}

function renderPool() {
  api('/candidates/pool?page=1&page_size=50').then(r => {
    if (r.code === 0 && r.data) {
      const list = r.data.list;
      document.getElementById('ct').innerHTML = `<div class="card">
        <div class="card-header"><h3 class="card-title">候选人池 (未关联职位)</h3><button class="btn btn-secondary" onclick="renderCandidates()">返回全部</button></div>
        ${list.length===0 ? '<div class="empty-state"><div class="empty-state-icon">📭</div>候选人池为空</div>' :
          `<table><thead><tr><th>姓名</th><th>邮箱</th><th>电话</th><th>当前公司</th><th>工作年限</th><th>期望薪资</th><th>来源</th><th>操作</th></tr></thead><tbody>
            ${list.map(c => `<tr><td><strong>${c.name}</strong></td><td>${c.email||'-'}</td><td>${c.phone||'-'}</td><td>${c.current_company||'-'}</td><td>${c.work_years}年</td><td>${c.expected_salary}K</td><td><span class="tag">${c.source}</span></td><td><button class="btn btn-sm btn-secondary" onclick="viewCand(${c.id})">详情</button> <button class="btn btn-sm btn-primary" onclick="assignJob(${c.id})">关联职位</button></td></tr>`).join('')}
          </tbody></table>`}
      </div>`;
    }
  });
}

function showCandForm(id) { if (id) api('/candidates/' + id).then(r => { if (r.code === 0) showCandFormModal(r.data); }); else showCandFormModal(null); }

function showCandFormModal(c) {
  showModal(`<div class="modal-header"><h3 class="modal-title">${c ? '编辑候选人' : '录入候选人'}</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
    <form onsubmit="saveCand(event, ${c?.id||'null'})">
      <div class="form-row"><div class="form-group"><label class="form-label">姓名 *</label><input class="form-input" id="cfn" value="${c?.name||''}" required></div><div class="form-group"><label class="form-label">邮箱</label><input type="email" class="form-input" id="cfe" value="${c?.email||''}"></div></div>
      <div class="form-row"><div class="form-group"><label class="form-label">电话</label><input class="form-input" id="cfp" value="${c?.phone||''}"></div><div class="form-group"><label class="form-label">当前公司</label><input class="form-input" id="cfc" value="${c?.current_company||''}"></div></div>
      <div class="form-row"><div class="form-group"><label class="form-label">工作年限</label><input type="number" class="form-input" id="cfy" value="${c?.work_years||0}"></div><div class="form-group"><label class="form-label">学历</label><input class="form-input" id="cfed" value="${c?.education||''}"></div></div>
      <div class="form-row"><div class="form-group"><label class="form-label">期望薪资(K)</label><input type="number" class="form-input" id="cfs" value="${c?.expected_salary||0}"></div><div class="form-group"><label class="form-label">来源</label>
        <select class="form-select" id="cfsrc"><option value="招聘网站" ${c?.source==='招聘网站'?'selected':''}>招聘网站</option><option value="内推" ${c?.source==='内推'?'selected':''}>内推</option><option value="猎头" ${c?.source==='猎头'?'selected':''}>猎头</option><option value="主动投递" ${!c||c.source==='主动投递'?'selected':''}>主动投递</option></select>
      </div></div>
      <div class="form-group"><label class="form-label">标签(逗号分隔)</label><input class="form-input" id="cftg" value="${c?.tags||''}"></div>
      <div class="form-group"><label class="form-label">备注</label><textarea class="form-textarea" id="cfrm" rows="3">${c?.remark||''}</textarea></div>
      <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">保存</button></div>
    </form>`);
}

function saveCand(e, id) {
  e.preventDefault();
  const d = { name: document.getElementById('cfn').value, email: document.getElementById('cfe').value, phone: document.getElementById('cfp').value, current_company: document.getElementById('cfc').value, work_years: parseInt(document.getElementById('cfy').value)||0, education: document.getElementById('cfed').value, expected_salary: parseInt(document.getElementById('cfs').value)||0, source: document.getElementById('cfsrc').value, tags: document.getElementById('cftg').value, remark: document.getElementById('cfrm').value };
  api(id ? '/candidates/' + id : '/candidates', { method: id ? 'PUT' : 'POST', body: JSON.stringify(d) }).then(r => {
    if (r.code === 0) { showToast('保存成功'); closeModal(); loadCands(); } else showToast(r.message, 'error');
  });
}

function viewCand(id) {
  api('/candidates/' + id).then(r => {
    if (r.code === 0) {
      const c = r.data;
      Promise.all([api('/candidates/' + id + '/notes'), api('/candidates/' + id + '/scores'), api('/jobs?page=1&page_size=100')]).then(([nr, sr, jr]) => {
        const notes = nr.code===0 ? (nr.data||[]) : [];
        const scores = sr.code===0 ? (sr.data||[]) : [];
        const avg = scores.length > 0 ? (scores.reduce((a,s)=>a+s.score,0)/scores.length).toFixed(1) : 0;
        const jobs = jr.code===0 ? (jr.data?.list||[]) : [];
        const isPDF = c.resume_path && c.resume_path.toLowerCase().endsWith('.pdf');
        const isWord = c.resume_path && (c.resume_path.toLowerCase().endsWith('.doc') || c.resume_path.toLowerCase().endsWith('.docx'));
        showModal(`<div class="modal-header"><h3 class="modal-title">${c.name}</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
          <div class="tabs"><div class="tab active" onclick="switchTab(event,'info')">基本信息</div><div class="tab" onclick="switchTab(event,'resume')">简历${c.resume_path?' ✓':''}</div><div class="tab" onclick="switchTab(event,'notes')">内部备注 (${notes.length})</div><div class="tab" onclick="switchTab(event,'scores')">评分 (${scores.length})</div></div>
          <div id="ti" style="line-height:1.8;"><p><strong>邮箱：</strong>${c.email||'-'}</p><p><strong>电话：</strong>${c.phone||'-'}</p><p><strong>当前公司：</strong>${c.current_company||'-'}</p><p><strong>工作年限：</strong>${c.work_years}年</p><p><strong>学历：</strong>${c.education||'-'}</p><p><strong>期望薪资：</strong>${c.expected_salary}K</p><p><strong>来源：</strong>${c.source}</p><p><strong>标签：</strong>${(c.tags||'').split(',').filter(t=>t).map(t=>`<span class="tag">${t}</span>`).join('')}</p><p><strong>备注：</strong>${c.remark||'-'}</p><p><strong>创建时间：</strong>${fmt(c.created_at)}</p></div>
          <div id="tr" style="display:none;">
            ${c.resume_path ?
              (isPDF ? `<div><p style="margin-bottom:10px;"><button class="btn btn-secondary btn-sm" onclick="window.open('/api/candidates/${id}/resume','_blank')">在新窗口打开</button></div><div class="resume-preview"><iframe src="/api/candidates/${id}/resume" style="width:100%;height:500px;border:1px solid #ddd;border-radius:6px;"></iframe></div>` :
               isWord ? `<div class="empty-state"><div class="empty-state-icon">📄</div><p>Word 格式简历，请下载后查看</p><button class="btn btn-primary" onclick="window.location.href='/api/candidates/${id}/resume'">下载简历</button></div>` :
               `<div class="empty-state"><div class="empty-state-icon">📁</div><p>简历格式不支持在线预览</p><button class="btn btn-primary" onclick="window.location.href='/api/candidates/${id}/resume'">下载简历</button></div>`) :
              `<div class="empty-state"><div class="empty-state-icon">📄</div><p>暂无简历上传</p><button class="btn btn-primary" onclick="closeModal();uploadResume(${id})">上传简历</button></div>`
            }
          </div>
          <div id="tn" style="display:none;"><div style="margin-bottom:15px;"><textarea class="form-textarea" id="nnc" placeholder="添加内部备注..." rows="3"></textarea><button class="btn btn-primary btn-sm" onclick="addNote(${id})">添加备注</button></div>
            ${notes.length===0 ? '<div style="color:#999;text-align:center;padding:20px;">暂无备注</div>' : notes.map(n => `<div style="padding:10px;border-bottom:1px solid #f0f0f0;"><div style="font-weight:500;">${n.user_name}</div><div style="margin:5px 0;">${n.content}</div><div style="font-size:11px;color:#999;">${fmt(n.created_at)}</div></div>`).join('')}
          </div>
          <div id="ts" style="display:none;"><div style="margin-bottom:15px;padding:10px;background:#f8f9fa;border-radius:6px;"><strong>平均评分：${avg} / 5</strong></div>
            <div style="margin-bottom:15px;"><div class="form-group"><label class="form-label">选择职位</label><select class="form-select" id="sj"><option value="">请选择</option>${jobs.map(j=>`<option value="${j.id}">${j.title}</option>`).join('')}</select></div>
              <div class="form-group"><label class="form-label">评分</label><select class="form-select" id="sv"><option value="5">5 - 强烈推荐</option><option value="4">4 - 推荐</option><option value="3">3 - 待定</option><option value="2">2 - 不推荐</option><option value="1">1 - 强烈不推荐</option></select></div>
              <div class="form-group"><label class="form-label">评语</label><textarea class="form-textarea" id="scm" rows="2"></textarea></div>
              <button class="btn btn-primary btn-sm" onclick="addScore(${id})">提交评分</button>
            </div>
            ${scores.length===0 ? '<div style="color:#999;text-align:center;padding:20px;">暂无评分</div>' : scores.map(s => `<div style="padding:10px;border-bottom:1px solid #f0f0f0;"><div style="font-weight:500;">${s.user_name} - ${s.score}分</div><div style="margin:5px 0;color:#666;">${s.comment||''}</div><div style="font-size:11px;color:#999;">${fmt(s.created_at)}</div></div>`).join('')}
          </div>`);
      });
    }
  });
}

function switchTab(e, t) {
  document.querySelectorAll('.tab').forEach(x => x.classList.remove('active'));
  e.target.classList.add('active');
  ['ti','tr','tn','ts'].forEach(id => {
    const el = document.getElementById(id);
    if (el) el.style.display = 'none';
  });
  const target = document.getElementById(t==='info'?'ti':t==='resume'?'tr':t==='notes'?'tn':'ts');
  if (target) target.style.display = 'block';
}

function addNote(id) { const v = document.getElementById('nnc').value; if (!v) return; api('/candidates/' + id + '/notes', { method: 'POST', body: JSON.stringify({ content: v }) }).then(r => { if (r.code === 0) { showToast('备注添加成功'); viewCand(id); } }); }
function addScore(id) { const j = document.getElementById('sj').value; if (!j) { showToast('请选择职位','error'); return; } api('/candidates/' + id + '/scores', { method: 'POST', body: JSON.stringify({ job_id: parseInt(j), score: parseInt(document.getElementById('sv').value), comment: document.getElementById('scm').value }) }).then(r => { if (r.code === 0) { showToast('评分成功'); viewCand(id); } }); }

function assignJob(id) {
  api('/jobs?page=1&page_size=100&status=open').then(r => {
    if (r.code === 0 && r.data?.list) {
      showModal(`<div class="modal-header"><h3 class="modal-title">关联到职位</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
        <form onsubmit="doAssign(event,${id})"><div class="form-group"><label class="form-label">选择职位 *</label><select class="form-select" id="aj" required><option value="">请选择</option>${r.data.list.map(j=>`<option value="${j.id}">${j.title} (${j.department})</option>`).join('')}</select></div>
          <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">确认关联</button></div>
        </form>`);
    }
  });
}

function doAssign(e, id) {
  e.preventDefault();
  api('/candidate-jobs', { method: 'POST', body: JSON.stringify({ candidate_id: id, job_id: parseInt(document.getElementById('aj').value) }) }).then(r => {
    if (r.code === 0) { showToast('关联成功'); closeModal(); navigate('workflow'); } else showToast(r.message, 'error');
  });
}

function renderWorkflow() {
  document.getElementById('ct').innerHTML = `<div class="card">
    <div class="card-header"><h3 class="card-title">招聘流程</h3><div class="btn-group"><button class="btn btn-secondary" onclick="showWfs()">流程配置</button></div></div>
    <div class="filter-bar"><div class="form-group"><label class="form-label">选择职位</label><select class="form-select" id="wj" onchange="loadKanban()"><option value="">请选择职位</option></select></div></div>
    <div id="kv"><div class="empty-state">请选择职位查看看板</div></div>
  </div>`;
  api('/jobs?page=1&page_size=100&status=open').then(r => { if (r.code===0 && r.data?.list) r.data.list.forEach(j => { document.getElementById('wj').innerHTML += `<option value="${j.id}">${j.title}</option>`; }); });
}

function loadKanban() {
  const jid = document.getElementById('wj').value;
  if (!jid) return;
  api('/candidate-jobs/kanban?job_id=' + jid).then(r => {
    if (r.code === 0) {
      const d = r.data || {};
      const stages = ['简历筛选','电话面试','技术面试','HR面试','Offer','入职'];
      document.getElementById('kv').innerHTML = `<div class="kanban-board">
        ${stages.map((n, i) => { const sn = i+1, cards = d[sn]||[]; return `<div class="kanban-column"><div class="kanban-column-header"><span>${n}</span><span class="status-badge status-draft">${cards.length}</span></div>${cards.map(c => `<div class="kanban-card"><div class="kanban-card-name">${c.name}</div><div class="kanban-card-info">${c.phone||c.email||''}</div><div class="kanban-card-info">${c.current_company||''} | ${c.work_years}年</div><div style="margin-top:8px;"><button class="btn btn-sm btn-secondary" onclick="showMoveStageModal(${c.candidate_job_id},${sn})">移动</button> <button class="btn btn-sm btn-danger" onclick="rejectCand(${c.candidate_job_id})">拒绝</button></div></div>`).join('')}</div>`; }).join('')}
      </div>`;
    }
  });
}

function showMoveStageModal(id, cur) {
  const stages = ['简历筛选','电话面试','技术面试','HR面试','Offer','入职'];
  showModal(`<div class="modal-header"><h3 class="modal-title">移动阶段</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
    <form onsubmit="doMoveStage(event,${id})">
      <div class="form-group"><label class="form-label">目标阶段</label>
        <select class="form-select" id="mss" required>${stages.map((n,i)=>`<option value="${i+1}" ${i+1===cur?'selected':''}>${n}</option>`).join('')}</select>
      </div>
      <div class="form-group"><label class="form-label">阶段变更评价</label><textarea class="form-textarea" id="mse" rows="3"></textarea></div>
      <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">确认移动</button></div>
    </form>`);
}

function doMoveStage(e, id) {
  e.preventDefault();
  const ns = parseInt(document.getElementById('mss').value);
  const ev = document.getElementById('mse').value;
  api('/candidate-jobs/' + id + '/move-stage', { method: 'PUT', body: JSON.stringify({ to_stage: ns, evaluation: ev }) }).then(r => {
    if (r.code === 0) { showToast('阶段更新成功'); closeModal(); loadKanban(); } else showToast(r.message, 'error');
  });
}

function rejectCand(id) {
  api('/candidate-jobs/rejection-reasons').then(r => {
    const rs = r.code===0 ? (r.data||[]) : [];
    showModal(`<div class="modal-header"><h3 class="modal-title">拒绝候选人</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
      <form onsubmit="doReject(event,${id})"><div class="form-group"><label class="form-label">拒绝原因 *</label><select class="form-select" id="rr" required><option value="">请选择</option>${rs.map(x=>`<option value="${x.reason}">${x.reason}</option>`).join('')}</select></div>
        <div class="form-group"><label class="form-label">是否发送感谢邮件</label><input type="checkbox" id="rrm"></div>
        <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-danger">确认拒绝</button></div>
      </form>`);
  });
}

function doReject(e, id) {
  e.preventDefault();
  api('/candidate-jobs/' + id + '/reject', { method: 'PUT', body: JSON.stringify({ reason: document.getElementById('rr').value, send_mail: document.getElementById('rrm').checked }) }).then(r => {
    if (r.code === 0) { showToast('已拒绝'); closeModal(); loadKanban(); }
  });
}

function showWfs() {
  api('/workflows').then(r => {
    const wfs = r.code===0 ? (r.data||[]) : [];
    showModal(`<div class="modal-header"><h3 class="modal-title">流程配置</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
      <div style="margin-bottom:15px;"><button class="btn btn-primary btn-sm" onclick="showWfForm()">+ 新建流程</button></div>
      ${wfs.map(w => `<div style="padding:12px;border:1px solid #eee;border-radius:6px;margin-bottom:10px;"><div style="display:flex;justify-content:space-between;align-items:center;"><div><strong>${w.name}</strong><p style="color:#888;font-size:12px;margin:4px 0;">${w.description||''}</p></div><div><button class="btn btn-sm btn-secondary" onclick="viewWf(${w.id})">查看</button> <button class="btn btn-sm btn-secondary" onclick="showWfForm(${w.id})">编辑</button> <button class="btn btn-sm btn-danger" onclick="delWf(${w.id})">删除</button></div></div></div>`).join('')}
      <div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`);
  });
}

function showWfForm(id) { if (id) api('/workflows/' + id).then(r => { if (r.code === 0) showWfFormModal(r.data); }); else showWfFormModal(null); }

function showWfFormModal(w) {
  showModal(`<div class="modal-header"><h3 class="modal-title">${w ? '编辑流程' : '新建流程'}</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
    <form onsubmit="saveWf(event, ${w?.id||'null'})">
      <div class="form-group"><label class="form-label">流程名称 *</label><input class="form-input" id="wfn" value="${w?.name||''}" required></div>
      <div class="form-group"><label class="form-label">描述</label><input class="form-input" id="wfd" value="${w?.description||''}"></div>
      <div class="form-group"><label class="form-label">阶段配置(JSON格式)</label><textarea class="form-textarea" id="wfs" rows="6">${w?.stages||'[{"id":1,"name":"简历筛选"},{"id":2,"name":"电话面试"},{"id":3,"name":"技术面试"},{"id":4,"name":"HR面试"},{"id":5,"name":"Offer"},{"id":6,"name":"入职"}]'}</textarea></div>
      <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">保存</button></div>
    </form>`);
}

function saveWf(e, id) {
  e.preventDefault();
  api(id ? '/workflows/' + id : '/workflows', { method: id ? 'PUT' : 'POST', body: JSON.stringify({ name: document.getElementById('wfn').value, description: document.getElementById('wfd').value, stages: document.getElementById('wfs').value }) }).then(r => {
    if (r.code === 0) { showToast('保存成功'); closeModal(); showWfs(); } else showToast(r.message, 'error');
  });
}

function viewWf(id) {
  api('/workflows/' + id).then(r => {
    if (r.code === 0) { const w = r.data; showModal(`<div class="modal-header"><h3 class="modal-title">${w.name}</h3><button class="modal-close" onclick="closeModal()">&times;</button></div><div style="line-height:1.8;"><p><strong>描述：</strong>${w.description||'-'}</p><p><strong>阶段配置：</strong></p><pre style="white-space:pre-wrap;background:#f8f9fa;padding:12px;border-radius:6px;">${w.stages||''}</pre></div><div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`); }
  });
}

function delWf(id) { if (!confirm('确定删除？')) return; api('/workflows/' + id, { method: 'DELETE' }).then(r => { if (r.code === 0) { showToast('删除成功'); showWfs(); } }); }

let itvViewMode = 'list';
let itvCalDate = new Date();

function renderInterviews() {
  api('/auth/users').then(ur => {
    const users = ur.code===0 ? (ur.data||[]) : [];
    document.getElementById('ct').innerHTML = `<div class="card">
      <div class="card-header"><h3 class="card-title">面试安排</h3>
        <div class="btn-group">
          <div class="view-toggle" style="margin-right:12px;">
            <button class="${itvViewMode==='list'?'active':''}" onclick="switchItvView('list')">列表</button>
            <button class="${itvViewMode==='calendar'?'active':''}" onclick="switchItvView('calendar')">日历</button>
          </div>
          <button class="btn btn-primary" onclick="showItvForm()">+ 创建面试</button>
        </div>
      </div>
      <div class="filter-bar">
        <div class="form-group"><label class="form-label">状态</label><select class="form-select" id="ist" onchange="refreshItvs()"><option value="">全部状态</option><option value="scheduled">待进行</option><option value="completed">已完成</option></select></div>
        <div class="form-group"><label class="form-label">面试官</label><select class="form-select" id="iir" onchange="refreshItvs()"><option value="">全部面试官</option>${users.map(u=>`<option value="${u.id}">${u.real_name}</option>`).join('')}</select></div>
        <div class="form-group"><label class="form-label">开始日期</label><input type="date" class="form-input" id="isd" onchange="refreshItvs()"></div>
        <div class="form-group"><label class="form-label">结束日期</label><input type="date" class="form-input" id="ied" onchange="refreshItvs()"></div>
      </div>
      <div id="il">加载中...</div>
    </div>`;
    refreshItvs();
  });
}

function switchItvView(view) {
  itvViewMode = view;
  renderInterviews();
}

function refreshItvs() {
  if (itvViewMode === 'calendar') loadCalendarEvents();
  else loadItvs();
}

function loadItvs() {
  let url = '/interviews?page_size=50';
  const st = document.getElementById('ist')?.value, sd = document.getElementById('isd')?.value, ed = document.getElementById('ied')?.value;
  if (st) url += `&status=${st}`; if (sd) url += `&start_date=${sd}`; if (ed) url += `&end_date=${ed}`;
  api(url).then(r => {
    if (r.code === 0 && r.data) {
      const list = r.data;
      document.getElementById('il').innerHTML = list.length===0 ? '<div class="empty-state"><div class="empty-state-icon">📅</div>暂无面试安排</div>' :
        `<table><thead><tr><th>候选人</th><th>职位</th><th>面试时间</th><th>方式</th><th>时长</th><th>状态</th><th>操作</th></tr></thead><tbody>
          ${list.map(i => `<tr><td><strong>${i.candidate_name||'-'}</strong></td><td>${i.job_title||'-'}</td><td>${fmt(i.interview_time)}</td><td><span class="tag">${i.method}</span></td><td>${i.duration}分钟</td><td><span class="status-badge status-${i.status}">${statusText(i.status)}</span></td><td><button class="btn btn-sm btn-secondary" onclick="viewItv(${i.id})">详情</button> <button class="btn btn-sm btn-secondary" onclick="showItvForm(${i.id})">编辑</button> ${i.status==='scheduled' ? `<button class="btn btn-sm btn-secondary" onclick="showEvalForm(${i.id})">评价</button>` : ''}</td></tr>`).join('')}
        </tbody></table>`;
    }
  });
}

function loadCalendarEvents() {
  const year = itvCalDate.getFullYear();
  const month = itvCalDate.getMonth();
  const firstDay = new Date(year, month, 1);
  const lastDay = new Date(year, month + 1, 0);
  const sd = firstDay.toISOString().split('T')[0];
  const ed = lastDay.toISOString().split('T')[0];

  let url = `/interviews/calendar?start_date=${sd}&end_date=${ed}`;
  const st = document.getElementById('ist')?.value;
  const ir = document.getElementById('iir')?.value;
  if (st) url += `&status=${st}`;
  if (ir) url += `&interviewer_id=${ir}`;

  api(url).then(r => {
    if (r.code === 0) {
      const events = r.data || [];
      renderCalendar(events, year, month);
    }
  });
}

function renderCalendar(events, year, month) {
  const monthNames = ['一月','二月','三月','四月','五月','六月','七月','八月','九月','十月','十一月','十二月'];
  const dayNames = ['日','一','二','三','四','五','六'];

  const firstDay = new Date(year, month, 1);
  const lastDay = new Date(year, month + 1, 0);
  const startDay = firstDay.getDay();
  const totalDays = lastDay.getDate();
  const prevLastDay = new Date(year, month, 0).getDate();

  const today = new Date();
  const todayStr = today.toDateString();

  let eventsByDate = {};
  events.forEach(e => {
    const d = e.interview_time.split('T')[0];
    if (!eventsByDate[d]) eventsByDate[d] = [];
    eventsByDate[d].push(e);
  });

  let html = `<div class="calendar-container">
    <div class="calendar-header">
      <button class="calendar-nav-btn" onclick="prevMonth()">&lsaquo;</button>
      <span class="calendar-title">${year}年 ${monthNames[month]}</span>
      <button class="calendar-nav-btn" onclick="nextMonth()">&rsaquo;</button>
    </div>
    <div class="calendar-grid">`;

  dayNames.forEach(d => { html += `<div class="calendar-day-header">${d}</div>`; });

  let cells = [];
  for (let i = 0; i < startDay; i++) {
    cells.push({ day: prevLastDay - startDay + 1 + i, otherMonth: true, date: `${year}-${String(month).padStart(2,'0')}-${String(prevLastDay - startDay + 1 + i).padStart(2,'0')}` });
  }
  for (let i = 1; i <= totalDays; i++) {
    cells.push({ day: i, otherMonth: false, date: `${year}-${String(month+1).padStart(2,'0')}-${String(i).padStart(2,'0')}` });
  }
  const remaining = 42 - cells.length;
  for (let i = 1; i <= remaining; i++) {
    cells.push({ day: i, otherMonth: true, date: `${year}-${String(month+2).padStart(2,'0')}-${String(i).padStart(2,'0')}` });
  }

  cells.forEach(cell => {
    const cellDate = new Date(cell.date);
    const isToday = cellDate.toDateString() === todayStr;
    const dayEvents = eventsByDate[cell.date] || [];
    const shownEvents = dayEvents.slice(0, 3);
    const moreCount = dayEvents.length - shownEvents.length;

    html += `<div class="calendar-day ${cell.otherMonth?'other-month':''} ${isToday?'today':''}">
      <div class="calendar-day-date">${cell.day}</div>`;

    shownEvents.forEach(e => {
      const timeStr = e.interview_time.split('T')[1]?.substring(0,5) || '';
      html += `<div class="calendar-event ${e.status==='completed'?'completed':''}" onclick="viewItv(${e.id})" title="${e.candidate_name} - ${e.job_title}\n${timeStr} ${e.method}">${timeStr} ${e.candidate_name||'-'}</div>`;
    });

    if (moreCount > 0) {
      html += `<div class="calendar-more" onclick="showDayEvents('${cell.date}')">+${moreCount} 更多</div>`;
    }

    html += `</div>`;
  });

  html += `</div></div>`;

  document.getElementById('il').innerHTML = html;
}

function prevMonth() {
  itvCalDate.setMonth(itvCalDate.getMonth() - 1);
  loadCalendarEvents();
}

function nextMonth() {
  itvCalDate.setMonth(itvCalDate.getMonth() + 1);
  loadCalendarEvents();
}

function showDayEvents(dateStr) {
  api(`/interviews/calendar?start_date=${dateStr}&end_date=${dateStr}`).then(r => {
    if (r.code === 0 && r.data) {
      const events = r.data;
      showModal(`<div class="modal-header"><h3 class="modal-title">${dateStr} 面试安排</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
        <div style="max-height:400px;overflow-y:auto;">
          ${events.length===0 ? '<div class="empty-state">暂无面试</div>' :
            events.map(e => `<div style="padding:12px;border:1px solid #eee;border-radius:6px;margin-bottom:10px;">
              <div style="display:flex;justify-content:space-between;align-items:center;">
                <div><strong>${e.candidate_name||'-'}</strong> - ${e.job_title||'-'}</div>
                <span class="status-badge status-${e.status}">${statusText(e.status)}</span>
              </div>
              <p style="margin:8px 0 0;font-size:13px;color:#666;">时间: ${fmt(e.interview_time)} | 方式: ${e.method} | 时长: ${e.duration}分钟</p>
              ${e.location ? `<p style="margin:4px 0 0;font-size:12px;color:#888;">地点: ${e.location}</p>` : ''}
              ${e.link ? `<p style="margin:4px 0 0;font-size:12px;color:#888;">链接: ${e.link}</p>` : ''}
              <div style="margin-top:8px;"><button class="btn btn-sm btn-secondary" onclick="viewItv(${e.id});closeModal();">查看详情</button></div>
            </div>`).join('')}
        </div>
        <div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`);
    }
  });
}

function showItvForm(id) {
  if (id) api('/interviews/' + id).then(r => { if (r.code === 0) showItvFormModal(r.data); });
  else {
    Promise.all([api('/candidate-jobs?status=active'), api('/auth/users')]).then(([cr, ur]) => {
      const cj = cr.code===0 ? (cr.data||[]) : [];
      const users = ur.code===0 ? (ur.data||[]) : [];
      showModal(`<div class="modal-header"><h3 class="modal-title">创建面试</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
        <form onsubmit="saveItv(event)">
          <div class="form-group"><label class="form-label">候选人-职位 *</label><select class="form-select" id="icj" required><option value="">请选择</option>${cj.map(x=>`<option value="${x.id}">${x.candidate_name} - ${x.job_title}</option>`).join('')}</select></div>
          <div class="form-group"><label class="form-label">面试官 * (按住Ctrl多选)</label><select class="form-select" id="iiu" multiple>${users.map(x=>`<option value="${x.id}">${x.real_name}</option>`).join('')}</select></div>
          <div class="form-row"><div class="form-group"><label class="form-label">面试时间 *</label><input type="datetime-local" class="form-input" id="itm" required></div><div class="form-group"><label class="form-label">时长(分钟)</label><input type="number" class="form-input" id="idu" value="60"></div></div>
          <div class="form-row"><div class="form-group"><label class="form-label">方式</label><select class="form-select" id="im"><option value="现场">现场</option><option value="视频">视频</option><option value="电话">电话</option></select></div><div class="form-group"><label class="form-label">地点</label><input class="form-input" id="ilc"></div></div>
          <div class="form-group"><label class="form-label">视频链接</label><input class="form-input" id="ilk"></div>
          <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">保存</button></div>
        </form>`);
    });
  }
}

function showItvFormModal(i) {
  showModal(`<div class="modal-header"><h3 class="modal-title">编辑面试</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
    <form onsubmit="saveItv(event, ${i.id})">
      <div class="form-row"><div class="form-group"><label class="form-label">面试时间 *</label><input type="datetime-local" class="form-input" id="itm" value="${i.interview_time?.substring(0,16)||''}" required></div><div class="form-group"><label class="form-label">时长(分钟)</label><input type="number" class="form-input" id="idu" value="${i.duration||60}"></div></div>
      <div class="form-row"><div class="form-group"><label class="form-label">方式</label><select class="form-select" id="im"><option value="现场" ${i.method==='现场'?'selected':''}>现场</option><option value="视频" ${i.method==='视频'?'selected':''}>视频</option><option value="电话" ${i.method==='电话'?'selected':''}>电话</option></select></div><div class="form-group"><label class="form-label">地点</label><input class="form-input" id="ilc" value="${i.location||''}"></div></div>
      <div class="form-group"><label class="form-label">视频链接</label><input class="form-input" id="ilk" value="${i.link||''}"></div>
      <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">保存</button></div>
    </form>`);
}

function saveItv(e, id) {
  e.preventDefault();
  const d = {
    candidate_job_id: id ? undefined : parseInt(document.getElementById('icj').value),
    interviewer_ids: id ? undefined : JSON.stringify(Array.from(document.getElementById('iiu').selectedOptions).map(o => parseInt(o.value))),
    interview_time: document.getElementById('itm').value,
    duration: parseInt(document.getElementById('idu').value) || 60,
    method: document.getElementById('im').value,
    location: document.getElementById('ilc').value,
    link: document.getElementById('ilk').value
  };
  api(id ? '/interviews/' + id : '/interviews', { method: id ? 'PUT' : 'POST', body: JSON.stringify(d) }).then(r => {
    if (r.code === 0) { showToast('保存成功'); closeModal(); loadItvs(); } else showToast(r.message, 'error');
  });
}

function viewItv(id) {
  api('/interviews/' + id).then(r => {
    if (r.code === 0) {
      const i = r.data;
      showModal(`<div class="modal-header"><h3 class="modal-title">面试详情</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
        <div style="line-height:1.8;"><p><strong>候选人：</strong>${i.candidate_name||'-'}</p><p><strong>职位：</strong>${i.job_title||'-'}</p><p><strong>面试时间：</strong>${fmt(i.interview_time)}</p><p><strong>方式：</strong>${i.method}</p><p><strong>时长：</strong>${i.duration}分钟</p><p><strong>地点：</strong>${i.location||'-'}</p><p><strong>链接：</strong>${i.link||'-'}</p><p><strong>状态：</strong><span class="status-badge status-${i.status}">${statusText(i.status)}</span></p>
          ${i.status==='completed' ? `<hr style="margin:15px 0;"><h4>面试评价</h4><p><strong>技术能力：</strong>${i.tech_score}/10</p><p><strong>沟通能力：</strong>${i.comm_score}/10</p><p><strong>文化匹配：</strong>${i.culture_score}/10</p><p><strong>综合评价：</strong>${i.overall_score}/10</p><p><strong>推荐级别：</strong>${i.recommendation||'-'}</p><pre style="white-space:pre-wrap;font-family:inherit;background:#f8f9fa;padding:12px;border-radius:6px;">${i.evaluation||'暂无评价'}</pre>` : ''}
        </div>
        <div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`);
    }
  });
}

function showEvalForm(id) {
  showModal(`<div class="modal-header"><h3 class="modal-title">面试评价</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
    <form onsubmit="saveEval(event, ${id})">
      <div class="form-row"><div class="form-group"><label class="form-label">技术能力(0-10)</label><input type="number" class="form-input" id="ets" min="0" max="10" value="0"></div><div class="form-group"><label class="form-label">沟通能力(0-10)</label><input type="number" class="form-input" id="ecs" min="0" max="10" value="0"></div></div>
      <div class="form-row"><div class="form-group"><label class="form-label">文化匹配(0-10)</label><input type="number" class="form-input" id="ecus" min="0" max="10" value="0"></div><div class="form-group"><label class="form-label">综合评价(0-10)</label><input type="number" class="form-input" id="eos" min="0" max="10" value="0"></div></div>
      <div class="form-group"><label class="form-label">推荐级别</label><select class="form-select" id="erc"><option value="强烈推荐">强烈推荐</option><option value="推荐">推荐</option><option value="待定">待定</option><option value="不推荐">不推荐</option></select></div>
      <div class="form-group"><label class="form-label">评语</label><textarea class="form-textarea" id="eev" rows="4"></textarea></div>
      <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">提交</button></div>
    </form>`);
}

function saveEval(e, id) {
  e.preventDefault();
  api('/interviews/' + id + '/evaluate', { method: 'PUT', body: JSON.stringify({
    evaluation: document.getElementById('eev').value, tech_score: parseInt(document.getElementById('ets').value)||0,
    comm_score: parseInt(document.getElementById('ecs').value)||0, culture_score: parseInt(document.getElementById('ecus').value)||0,
    overall_score: parseInt(document.getElementById('eos').value)||0, recommendation: document.getElementById('erc').value
  }) }).then(r => {
    if (r.code === 0) { showToast('评价提交成功'); closeModal(); loadItvs(); } else showToast(r.message, 'error');
  });
}

function renderOffers() {
  document.getElementById('ct').innerHTML = `<div class="card">
    <div class="card-header"><h3 class="card-title">Offer管理</h3>
      <div class="btn-group"><button class="btn btn-secondary" onclick="showOfferTpls()">Offer模板</button><button class="btn btn-primary" onclick="showOfferForm()">+ 创建Offer</button></div>
    </div>
    <div class="filter-bar">
      <div class="form-group"><label class="form-label">状态</label><select class="form-select" id="ost" onchange="loadOffers()"><option value="">全部状态</option><option value="pending_approval">待审批</option><option value="approved">已审批</option><option value="sent">已发送</option><option value="accepted">已接受</option><option value="rejected_offer">已拒绝</option><option value="expired">已过期</option></select></div>
    </div>
    <div id="ol">加载中...</div>
  </div>`;
  loadOffers();
}

function loadOffers() {
  let url = '/offers?page_size=50';
  const st = document.getElementById('ost')?.value;
  if (st) url += `&status=${st}`;
  api(url).then(r => {
    if (r.code === 0 && r.data) {
      const list = r.data;
      document.getElementById('ol').innerHTML = list.length===0 ? '<div class="empty-state"><div class="empty-state-icon">📄</div>暂无Offer</div>' :
        `<table><thead><tr><th>候选人</th><th>职位</th><th>薪资</th><th>入职日期</th><th>状态</th><th>创建时间</th><th>操作</th></tr></thead><tbody>
          ${list.map(o => `<tr><td><strong>${o.candidate_name||'-'}</strong></td><td>${o.job_title||'-'}</td><td>${o.salary}K</td><td>${o.start_date||'-'}</td><td><span class="status-badge status-${o.status}">${statusText(o.status)}</span></td><td>${fmt(o.created_at)}</td><td><button class="btn btn-sm btn-secondary" onclick="viewOffer(${o.id})">详情</button> <button class="btn btn-sm btn-secondary" onclick="showOfferForm(${o.id})">编辑</button> ${o.status==='pending_approval' ? `<button class="btn btn-sm btn-primary" onclick="showApproval(${o.id})">审批</button>` : ''} ${o.status==='approved' ? `<button class="btn btn-sm btn-primary" onclick="sendOffer(${o.id})">发送</button>` : ''}</td></tr>`).join('')}
        </tbody></table>`;
    }
  });
}

function showOfferForm(id) {
  if (id) api('/offers/' + id).then(r => { if (r.code === 0) showOfferFormModal(r.data); });
  else {
    api('/candidate-jobs?status=active').then(r => {
      const cj = r.code===0 ? (r.data||[]) : [];
      showModal(`<div class="modal-header"><h3 class="modal-title">创建Offer</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
        <form onsubmit="saveOffer(event)">
          <div class="form-group"><label class="form-label">候选人-职位 *</label><select class="form-select" id="ocj" required><option value="">请选择</option>${cj.map(x=>`<option value="${x.id}">${x.candidate_name} - ${x.job_title}</option>`).join('')}</select></div>
          <div class="form-row"><div class="form-group"><label class="form-label">薪资(K) *</label><input type="number" class="form-input" id="os" required></div><div class="form-group"><label class="form-label">入职日期 *</label><input type="date" class="form-input" id="osd" required></div></div>
          <div class="form-group"><label class="form-label">其他条款</label><textarea class="form-textarea" id="ot" rows="3"></textarea></div>
          <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">保存</button></div>
        </form>`);
    });
  }
}

function showOfferFormModal(o) {
  showModal(`<div class="modal-header"><h3 class="modal-title">编辑Offer</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
    <form onsubmit="saveOffer(event, ${o.id})">
      <div class="form-row"><div class="form-group"><label class="form-label">薪资(K) *</label><input type="number" class="form-input" id="os" value="${o.salary}" required></div><div class="form-group"><label class="form-label">入职日期 *</label><input type="date" class="form-input" id="osd" value="${o.start_date||''}" required></div></div>
      <div class="form-group"><label class="form-label">其他条款</label><textarea class="form-textarea" id="ot" rows="3">${o.terms||''}</textarea></div>
      <div class="form-group"><label class="form-label">状态</label>
        <select class="form-select" id="ost"><option value="pending_approval" ${o.status==='pending_approval'?'selected':''}>待审批</option><option value="approved" ${o.status==='approved'?'selected':''}>已审批</option><option value="sent" ${o.status==='sent'?'selected':''}>已发送</option><option value="accepted" ${o.status==='accepted'?'selected':''}>已接受</option><option value="rejected_offer" ${o.status==='rejected_offer'?'selected':''}>已拒绝</option><option value="expired" ${o.status==='expired'?'selected':''}>已过期</option></select>
      </div>
      <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">保存</button></div>
    </form>`);
}

function saveOffer(e, id) {
  e.preventDefault();
  const d = {
    candidate_job_id: id ? undefined : parseInt(document.getElementById('ocj').value),
    salary: parseInt(document.getElementById('os').value)||0,
    start_date: document.getElementById('osd').value,
    terms: document.getElementById('ot').value,
    status: id ? document.getElementById('ost').value : undefined
  };
  api(id ? '/offers/' + id : '/offers', { method: id ? 'PUT' : 'POST', body: JSON.stringify(d) }).then(r => {
    if (r.code === 0) { showToast('保存成功'); closeModal(); loadOffers(); } else showToast(r.message, 'error');
  });
}

function viewOffer(id) {
  api('/offers/' + id).then(r => {
    if (r.code === 0) {
      const o = r.data;
      api('/offers/' + id + '/approval').then(ar => {
        const appr = ar.code===0 ? (ar.data||[]) : [];
        showModal(`<div class="modal-header"><h3 class="modal-title">Offer详情</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
          <div style="line-height:1.8;"><p><strong>候选人：</strong>${o.candidate_name||'-'}</p><p><strong>职位：</strong>${o.job_title||'-'}</p><p><strong>薪资：</strong>${o.salary}K/月</p><p><strong>入职日期：</strong>${o.start_date||'-'}</p><p><strong>条款：</strong>${o.terms||'-'}</p><p><strong>状态：</strong><span class="status-badge status-${o.status}">${statusText(o.status)}</span></p><p><strong>创建时间：</strong>${fmt(o.created_at)}</p>
            ${appr.length>0 ? `<hr style="margin:15px 0;"><h4>审批进度</h4>`+appr.map(a=>`<div style="padding:8px;border-left:3px solid ${a.status==='approved'?'#48bb78':a.status==='rejected'?'#e53e3e':'#888'};margin-bottom:8px;padding-left:10px;"><strong>${a.approver_name||'-'}</strong> - <span class="status-badge status-${a.status}">${statusText(a.status)}</span>${a.comment?`<p style="margin:4px 0 0;font-size:12px;">${a.comment}</p>`:''}${a.approved_at?`<p style="margin:4px 0 0;font-size:11px;color:#888;">${fmt(a.approved_at)}</p>`:''}</div>`).join('') : ''}
          </div>
          <div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`);
      });
    }
  });
}

function showApproval(id) {
  api('/auth/users').then(r => {
    const users = r.code===0 ? (r.data||[]) : [];
    showModal(`<div class="modal-header"><h3 class="modal-title">设置审批链</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
      <form onsubmit="saveApproval(event, ${id})">
        <div class="form-group"><label class="form-label">审批人 (按住Ctrl选择多个)</label><select class="form-select" id="app" multiple>${users.map(u=>`<option value="${u.id}">${u.real_name} (${u.role})</option>`).join('')}</select></div>
        <div style="font-size:12px;color:#888;">审批将按选择顺序依次进行</div>
        <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">提交</button></div>
      </form>`);
  });
}

function saveApproval(e, id) {
  e.preventDefault();
  const ids = Array.from(document.getElementById('app').selectedOptions).map(o => parseInt(o.value));
  if (ids.length === 0) { showToast('请选择审批人', 'error'); return; }
  api('/offers/' + id + '/approval', { method: 'POST', body: JSON.stringify({ approver_ids: ids }) }).then(r => {
    if (r.code === 0) { showToast('审批链已设置'); closeModal(); loadOffers(); } else showToast(r.message, 'error');
  });
}

function sendOffer(id) {
  if (!confirm('确认发送Offer？系统将生成PDF并更新状态。')) return;
  api('/offers/' + id + '/send', { method: 'POST' }).then(r => {
    if (r.code === 0) { showToast('Offer已发送'); loadOffers(); } else showToast(r.message, 'error');
  });
}

function showOfferTpls() {
  api('/offers/templates').then(r => {
    const ts = r.code===0 ? (r.data||[]) : [];
    showModal(`<div class="modal-header"><h3 class="modal-title">Offer模板</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
      <div style="margin-bottom:15px;"><button class="btn btn-primary btn-sm" onclick="showOfferTplForm()">+ 新建模板</button></div>
      ${ts.length===0 ? '<div class="empty-state">暂无模板</div>' : ts.map(t=>`<div style="padding:12px;border:1px solid #eee;border-radius:6px;margin-bottom:10px;"><div style="display:flex;justify-content:space-between;align-items:center;"><div><strong>${t.name}</strong></div><div><button class="btn btn-sm btn-danger" onclick="delOfferTpl(${t.id})">删除</button></div></div></div>`).join('')}
      <div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`);
  });
}

function showOfferTplForm() {
  showModal(`<div class="modal-header"><h3 class="modal-title">新建模板</h3><button class="modal-close" onclick="closeModal()">&times;</button></div>
    <form onsubmit="saveOfferTpl(event)">
      <div class="form-group"><label class="form-label">模板名称 *</label><input class="form-input" id="otn" required></div>
      <div class="form-group"><label class="form-label">薪资结构</label><input class="form-input" id="otss"></div>
      <div class="form-group"><label class="form-label">条款</label><textarea class="form-textarea" id="ott" rows="3"></textarea></div>
      <div class="modal-footer"><button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button><button type="submit" class="btn btn-primary">保存</button></div>
    </form>`);
}

function saveOfferTpl(e) {
  e.preventDefault();
  api('/offers/templates', { method: 'POST', body: JSON.stringify({ name: document.getElementById('otn').value, salary_structure: document.getElementById('otss').value, terms: document.getElementById('ott').value }) }).then(r => {
    if (r.code === 0) { showToast('模板创建成功'); closeModal(); } else showToast(r.message, 'error');
  });
}

function delOfferTpl(id) {
  if (!confirm('确定删除？')) return;
  api('/offers/templates/' + id, { method: 'DELETE' }).then(r => { if (r.code === 0) { showToast('删除成功'); showOfferTpls(); } });
}

function renderAnalytics() {
  Promise.all([api('/analytics/funnel'), api('/analytics/channels'), api('/analytics/interviewers'), api('/analytics/offer-acceptance'), api('/analytics/monthly-trend')]).then(([f, ch, iv, oa, mt]) => {
    const funnel = f.data?.stages || [];
    const channels = ch.data || [];
    const offerAccept = oa.data || {};
    const trend = mt.data || [];

    document.getElementById('ct').innerHTML = `<div class="card"><div class="card-header"><h3 class="card-title">招聘漏斗</h3></div>
      <div style="position:relative;height:300px;"><canvas id="funnelChart"></canvas></div>
    </div>
    <div class="card"><div class="card-header"><h3 class="card-title">月度趋势</h3></div>
      <div style="position:relative;height:300px;"><canvas id="trendChart"></canvas></div>
    </div>
    <div class="card"><div class="card-header"><h3 class="card-title">渠道效果</h3></div>
      <div style="position:relative;height:300px;"><canvas id="channelChart"></canvas></div>
    </div>
    <div class="card"><div class="card-header"><h3 class="card-title">Offer接受率</h3></div>
      <div style="position:relative;height:300px;display:flex;align-items:center;justify-content:center;"><div style="width:50%;"><canvas id="offerChart"></canvas></div><div style="width:50%;text-align:center;"><div style="font-size:48px;font-weight:bold;color:#667eea;">${(offerAccept.acceptance_rate||0).toFixed(1)}%</div><div style="color:#888;margin-top:10px;">已发送 ${offerAccept.total_sent||0} 个，已接受 ${offerAccept.total_accepted||0} 个</div></div></div>
    </div>
    <div class="card"><div class="card-header"><h3 class="card-title">面试官工作量</h3></div>
      <table><thead><tr><th>面试官</th><th>部门</th><th>面试数</th><th>平均分</th></tr></thead><tbody>
        ${(iv.data||[]).map(x => `<tr><td>${x.real_name}</td><td>${x.department||'-'}</td><td>${x.interview_count}</td><td>${x.avg_score.toFixed(1)}</td></tr>`).join('')}
      </tbody></table>
    </div>`;

    renderFunnelChart(funnel);
    renderTrendChart(trend);
    renderChannelChart(channels);
    renderOfferChart(offerAccept);
  });
}

function renderFunnelChart(data) {
  const ctx = document.getElementById('funnelChart');
  if (!ctx || typeof Chart === 'undefined') return;
  new Chart(ctx, {
    type: 'bar',
    data: {
      labels: data.map(d => d.stage_name),
      datasets: [{
        label: '人数',
        data: data.map(d => d.count),
        backgroundColor: [
          'rgba(102, 126, 234, 0.9)',
          'rgba(118, 75, 162, 0.85)',
          'rgba(159, 122, 234, 0.8)',
          'rgba(236, 72, 153, 0.75)',
          'rgba(246, 109, 155, 0.7)',
          'rgba(249, 168, 212, 0.65)'
        ],
        borderRadius: 4
      }]
    },
    options: {
      indexAxis: 'y',
      plugins: {
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: ctx => {
              const d = data[ctx.dataIndex];
              return `${d.count}人 (转化率: ${d.rate.toFixed(1)}%)`;
            }
          }
        }
      },
      scales: {
        x: { beginAtZero: true, grid: { color: 'rgba(0,0,0,0.05)' } },
        y: { grid: { display: false } }
      }
    }
  });
}

function renderTrendChart(data) {
  const ctx = document.getElementById('trendChart');
  if (!ctx || typeof Chart === 'undefined') return;
  new Chart(ctx, {
    type: 'line',
    data: {
      labels: data.map(d => d.month),
      datasets: [
        { label: '新增候选人', data: data.map(d => d.new_candidates), borderColor: '#667eea', backgroundColor: 'rgba(102, 126, 234, 0.1)', fill: true, tension: 0.3 },
        { label: '面试', data: data.map(d => d.interviews), borderColor: '#f6ad55', backgroundColor: 'rgba(246, 173, 85, 0.1)', fill: true, tension: 0.3 },
        { label: 'Offer', data: data.map(d => d.offers), borderColor: '#48bb78', backgroundColor: 'rgba(72, 187, 120, 0.1)', fill: true, tension: 0.3 },
        { label: '入职', data: data.map(d => d.hired), borderColor: '#ed64a6', backgroundColor: 'rgba(237, 100, 166, 0.1)', fill: true, tension: 0.3 }
      ]
    },
    options: {
      plugins: { legend: { position: 'bottom' } },
      scales: {
        x: { grid: { color: 'rgba(0,0,0,0.05)' } },
        y: { beginAtZero: true, grid: { color: 'rgba(0,0,0,0.05)' } }
      }
    }
  });
}

function renderChannelChart(data) {
  const ctx = document.getElementById('channelChart');
  if (!ctx || typeof Chart === 'undefined') return;
  new Chart(ctx, {
    type: 'bar',
    data: {
      labels: data.map(d => d.source),
      datasets: [
        { label: '总数', data: data.map(d => d.total_count), backgroundColor: 'rgba(102, 126, 234, 0.8)' },
        { label: '活跃', data: data.map(d => d.active_count), backgroundColor: 'rgba(72, 187, 120, 0.8)' },
        { label: 'Offer', data: data.map(d => d.offer_count), backgroundColor: 'rgba(246, 173, 85, 0.8)' },
        { label: '入职', data: data.map(d => d.hired_count), backgroundColor: 'rgba(237, 100, 166, 0.8)' }
      ]
    },
    options: {
      plugins: { legend: { position: 'bottom' } },
      scales: {
        x: { grid: { display: false } },
        y: { beginAtZero: true, grid: { color: 'rgba(0,0,0,0.05)' } }
      }
    }
  });
}

function renderOfferChart(data) {
  const ctx = document.getElementById('offerChart');
  if (!ctx || typeof Chart === 'undefined') return;
  const accepted = data.total_accepted || 0;
  const rejected = (data.total_sent || 0) - accepted;
  new Chart(ctx, {
    type: 'doughnut',
    data: {
      labels: ['已接受', '已拒绝/待确认'],
      datasets: [{
        data: [accepted, rejected],
        backgroundColor: ['#48bb78', '#e2e8f0'],
        borderWidth: 0
      }]
    },
    options: {
      cutout: '65%',
      plugins: { legend: { position: 'bottom' } }
    }
  });
}

function renderSettings() {
  Promise.all([api('/settings')]).then(([sr]) => {
    const smtp = sr.code===0 ? sr.data : {};
    document.getElementById('ct').innerHTML = `<div class="card">
    <div class="card-header"><h3 class="card-title">个人设置</h3></div>
    <form onsubmit="updateUser(event)">
      <div class="form-row"><div class="form-group"><label class="form-label">姓名</label><input class="form-input" id="srn" value="${currentUser?.real_name||''}"></div><div class="form-group"><label class="form-label">部门</label><input class="form-input" id="sd" value="${currentUser?.department||''}"></div></div>
      <div class="modal-footer" style="padding:0;"><button type="submit" class="btn btn-primary">保存</button></div>
    </form>
  </div>
  <div class="card">
    <div class="card-header"><h3 class="card-title">修改密码</h3></div>
    <form onsubmit="changePwd(event)">
      <div class="form-group"><label class="form-label">原密码</label><input type="password" class="form-input" id="op" required></div>
      <div class="form-group"><label class="form-label">新密码(至少6位)</label><input type="password" class="form-input" id="np" required></div>
      <div class="modal-footer" style="padding:0;"><button type="submit" class="btn btn-primary">修改</button></div>
    </form>
  </div>
  <div class="card">
    <div class="card-header"><h3 class="card-title">SMTP 邮件配置</h3></div>
    <form onsubmit="saveSMTPSettings(event)">
      <div class="form-row"><div class="form-group"><label class="form-label">SMTP 服务器</label><input class="form-input" id="smtp_host" value="${smtp.smtp_host||''}" placeholder="smtp.example.com"></div><div class="form-group"><label class="form-label">端口</label><input class="form-input" id="smtp_port" value="${smtp.smtp_port||'587'}" placeholder="587"></div></div>
      <div class="form-row"><div class="form-group"><label class="form-label">用户名</label><input class="form-input" id="smtp_user" value="${smtp.smtp_user||''}" placeholder="your@email.com"></div><div class="form-group"><label class="form-label">密码</label><input type="password" class="form-input" id="smtp_pass" value="${smtp.smtp_pass||''}" placeholder="password or app password"></div></div>
      <div class="form-row"><div class="form-group"><label class="form-label">发件人名称</label><input class="form-input" id="smtp_from" value="${smtp.smtp_from||''}" placeholder="HR <hr@company.com>"></div><div class="form-group"><label class="form-label">安全协议</label><select class="form-select" id="smtp_security"><option value="tls" ${smtp.smtp_security==='tls'?'selected':''}>TLS</option><option value="ssl" ${smtp.smtp_security==='ssl'?'selected':''}>SSL</option><option value="none" ${smtp.smtp_security==='none'?'selected':''}>无</option></select></div></div>
      <div class="modal-footer" style="padding:0;"><div style="display:flex;gap:8px;"><input type="email" class="form-input" id="test_email" style="width:200px;" placeholder="测试邮件地址"><button type="button" class="btn btn-secondary" onclick="testSMTP()">测试连接</button></div><button type="submit" class="btn btn-primary">保存配置</button></div>
    </form>
  </div>
  <div class="card">
    <div class="card-header"><h3 class="card-title">用户管理</h3></div>
    <div id="ul">加载中...</div>
  </div>
  <div class="card">
    <div class="card-header"><h3 class="card-title">周报</h3></div>
    <div class="btn-group" style="margin-bottom:15px;"><button class="btn btn-primary" onclick="genWeeklyReport()">生成本周周报</button></div>
    <div id="wrl">加载中...</div>
  </div>`;
  loadUserList(); loadWeeklyReports();
  });
}

function saveSMTPSettings(e) {
  e.preventDefault();
  const d = {
    smtp_host: document.getElementById('smtp_host').value,
    smtp_port: document.getElementById('smtp_port').value,
    smtp_user: document.getElementById('smtp_user').value,
    smtp_pass: document.getElementById('smtp_pass').value,
    smtp_from: document.getElementById('smtp_from').value,
    smtp_security: document.getElementById('smtp_security').value
  };
  api('/settings', { method: 'POST', body: JSON.stringify(d) }).then(r => {
    if (r.code === 0) showToast('SMTP配置保存成功');
    else showToast(r.message, 'error');
  });
}

function testSMTP() {
  const d = {
    smtp_host: document.getElementById('smtp_host').value,
    smtp_port: document.getElementById('smtp_port').value,
    smtp_user: document.getElementById('smtp_user').value,
    smtp_pass: document.getElementById('smtp_pass').value,
    smtp_from: document.getElementById('smtp_from').value,
    smtp_security: document.getElementById('smtp_security').value,
    test_email: document.getElementById('test_email').value
  };
  if (!d.test_email) { showToast('请输入测试邮件地址','error'); return; }
  if (!d.smtp_host || !d.smtp_user || !d.smtp_pass) { showToast('请先填写完整的SMTP配置','error'); return; }
  api('/settings/test-smtp', { method: 'POST', body: JSON.stringify(d) }).then(r => {
    if (r.code === 0) showToast('测试邮件发送成功！请查收。');
    else showToast(r.message, 'error');
  });
}

function updateUser(e) {
  e.preventDefault();
  api('/auth/user', { method: 'PUT', body: JSON.stringify({ real_name: document.getElementById('srn').value, department: document.getElementById('sd').value }) }).then(r => {
    if (r.code === 0) { showToast('保存成功'); currentUser.real_name = document.getElementById('srn').value; updateUser(); }
    else showToast(r.message, 'error');
  });
}

function changePwd(e) {
  e.preventDefault();
  api('/auth/password', { method: 'PUT', body: JSON.stringify({ old_password: document.getElementById('op').value, new_password: document.getElementById('np').value }) }).then(r => {
    if (r.code === 0) { showToast('密码修改成功'); } else showToast(r.message, 'error');
  });
}

function loadUserList() {
  api('/auth/users').then(r => {
    if (r.code === 0 && r.data) {
      document.getElementById('ul').innerHTML = `<table><thead><tr><th>用户名</th><th>姓名</th><th>邮箱</th><th>部门</th><th>角色</th><th>状态</th></tr></thead><tbody>
        ${r.data.map(u=>`<tr><td>${u.username}</td><td>${u.real_name}</td><td>${u.email}</td><td>${u.department||'-'}</td><td>${u.role}</td><td><span class="status-badge status-${u.status}">${statusText(u.status)}</span></td></tr>`).join('')}
      </tbody></table>`;
    }
  });
}

function genWeeklyReport() {
  api('/reports/weekly', { method: 'POST' }).then(r => {
    if (r.code === 0) { showToast('周报生成成功'); loadWeeklyReports(); } else showToast(r.message, 'error');
  });
}

function loadWeeklyReports() {
  api('/reports/weekly').then(r => {
    if (r.code === 0 && r.data) {
      document.getElementById('wrl').innerHTML = r.data.length===0 ? '<div class="empty-state">暂无周报</div>' :
        r.data.map(r=>`<div style="padding:12px;border:1px solid #eee;border-radius:6px;margin-bottom:10px;"><div style="font-weight:500;margin-bottom:6px;">${r.report_date}</div><pre style="white-space:pre-wrap;font-family:inherit;background:#f8f9fa;padding:12px;border-radius:6px;font-size:13px;">${r.content}</pre></div>`).join('');
    }
  });
}

if (token) initApp(); else showLoginPage();
