let page = 0;
const limit = 10;
let loading = false;
let hasMore = true;

const pathParts = window.location.pathname.split('/');
const pageOwnerUsername = pathParts[pathParts.length - 1];

async function loadUserActivities() {
 
  if (loading || !hasMore) return;
  loading = true;

  try {
    const res = await fetch(`/useractivity?limit=${limit}&page=${page}&username=${pageOwnerUsername}`);
    if (!res.ok) throw new Error("Failed to fetch activity");
    const activity = await res.json();

    if (Array.isArray(activity) && activity.length > 0) {
      formatUserActivities(activity);
    
      page++;
    } else {
      hasMore = false;
    }
  } catch (err) {
    console.error("Error loading user activities:", err);
  }

  loading = false;
}


function dh(html) {
  const entities = {
    '&#39;': "'",
    '&lt;': '<',
    '&gt;': '>',
    '&amp;': '&',
    // add more if needed
  };
  return html.replace(/&#39;|&lt;|&gt;|&amp;/g, match => entities[match]);
}

function formatUserActivities(activity) {
  const container = document.getElementById('user-activity');

  activity.forEach(a => {
    const created = a.created ? new Date(a.created).toLocaleString() : 'Unknown time';

    const hasPost = a.postId?.Valid;
    const hasComment = a.commentId?.Valid;
    const hasReaction = a.reactionId?.Valid;
    const preview = escapeHtml(a.preview)
    const bonus = escapeHtml(a.bonusText)

    let target = '';
    let action = '';
    let bonusText = bonus.charAt(0).toUpperCase() + bonus.slice(1);
    
    if (hasComment && hasPost && !hasReaction) {
      action = 'Comment'
      target = dh(`on post <a href="/post/${a.postId.Int64}">${preview}</a>`);
    } else if (hasPost && !hasReaction) {
      action = 'Post'
      target = dh(`<a href="/post/${a.postId.Int64}">${preview}</a>`);
    } else if (hasComment && hasReaction) {
      action = a.actionType.charAt(0).toUpperCase() + a.actionType.slice(1);
      target = dh(`on comment on post <a href="/post/${a.postId.Int64}">${preview}</a>`);
    } else if (hasPost && hasReaction) {
      action = a.actionType.charAt(0).toUpperCase() + a.actionType.slice(1);
      target = dh(`on post <a href="/post/${a.postId.Int64}">${preview}</a>`);
    } else if (hasComment && hasReaction) {
      action = a.actionType.charAt(0).toUpperCase() + a.actionType.slice(1);
      target = dh(`on comment ${preview} on post <a href="/post/${a.postId.Int64}">${bonus}</a>`);
    } else {
      action = a.actionType.charAt(0).toUpperCase() + a.actionType.slice(1);
      bonusText = preview
    }
    console.log(target)

    const div = document.createElement('div');
    div.style.marginBottom = '0.5em';
    div.innerHTML = `
      <p class="activity">
      <strong>${action}</strong>
      <span class="target-text">${target}</span><br>
      <span class="timestamp">${bonusText}</span><br>
      <small class="timestamp">${created}</small><br>
      <p>
    `;

    container.appendChild(div);
  });
}

function escapeHtml(str) {
  return String(str).replace(/[&<>"']/g, m => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  })[m]);
}

// Trigger initial load
document.addEventListener("DOMContentLoaded", () => {
  loadUserActivities();
});

// Infinite scroll trigger
window.addEventListener('scroll', () => {
  const scrollPos = window.scrollY + window.innerHeight;
  const docHeight = document.documentElement.offsetHeight;

  if (scrollPos > docHeight - 100) {
    loadUserActivities();
  }
});