// SublimeGo Dashboard - Main Application JavaScript
// Version: 2.0 - Custom Dashboard

// ============================================
// UTILITY FUNCTIONS
// ============================================
const Utils = {
    // Format number with thousands separator
    formatNumber(num) {
        return new Intl.NumberFormat('fr-FR').format(num);
    },

    // Format currency
    formatCurrency(num, currency = 'EUR') {
        return new Intl.NumberFormat('fr-FR', {
            style: 'currency',
            currency: currency
        }).format(num);
    },

    // Format date
    formatDate(date, options = {}) {
        const defaultOptions = {
            day: '2-digit',
            month: 'short',
            year: 'numeric'
        };
        return new Intl.DateTimeFormat('fr-FR', { ...defaultOptions, ...options }).format(new Date(date));
    },

    // Format datetime
    formatDateTime(date) {
        return new Intl.DateTimeFormat('fr-FR', {
            day: '2-digit',
            month: 'short',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        }).format(new Date(date));
    },

    // Time ago (relative time)
    timeAgo(date) {
        const seconds = Math.floor((new Date() - new Date(date)) / 1000);
        const intervals = {
            année: 31536000,
            mois: 2592000,
            semaine: 604800,
            jour: 86400,
            heure: 3600,
            minute: 60
        };
        for (const [unit, secondsInUnit] of Object.entries(intervals)) {
            const interval = Math.floor(seconds / secondsInUnit);
            if (interval >= 1) {
                return `Il y a ${interval} ${unit}${interval > 1 && unit !== 'mois' ? 's' : ''}`;
            }
        }
        return 'À l\'instant';
    },

    // Debounce function
    debounce(func, wait = 300) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    },

    // Throttle function
    throttle(func, limit = 100) {
        let inThrottle;
        return function(...args) {
            if (!inThrottle) {
                func.apply(this, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    },

    // Generate unique ID
    uniqueId(prefix = 'id') {
        return `${prefix}_${Math.random().toString(36).substr(2, 9)}`;
    },

    // Deep clone object
    deepClone(obj) {
        return JSON.parse(JSON.stringify(obj));
    },

    // Check if element is in viewport
    isInViewport(element) {
        const rect = element.getBoundingClientRect();
        return (
            rect.top >= 0 &&
            rect.left >= 0 &&
            rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
            rect.right <= (window.innerWidth || document.documentElement.clientWidth)
        );
    },

    // Escape HTML
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
};

// ============================================
// DATA TABLE - Enhanced Table Functionality
// ============================================
const DataTable = {
    instances: new Map(),

    init(tableId, options = {}) {
        const table = document.getElementById(tableId);
        if (!table) return null;

        const instance = {
            table,
            options: {
                searchable: true,
                sortable: true,
                pagination: true,
                perPage: 10,
                currentPage: 1,
                ...options
            },
            originalRows: [],
            filteredRows: []
        };

        // Store original rows
        instance.originalRows = Array.from(table.querySelectorAll('tbody tr'));
        instance.filteredRows = [...instance.originalRows];

        if (instance.options.searchable) this.initSearch(instance);
        if (instance.options.sortable) this.initSort(instance);
        if (instance.options.pagination) this.initPagination(instance);

        this.instances.set(tableId, instance);
        return instance;
    },

    initSearch(instance) {
        const searchInput = document.querySelector(`[data-table-search="${instance.table.id}"]`);
        if (!searchInput) return;

        searchInput.addEventListener('input', Utils.debounce((e) => {
            const searchTerm = e.target.value.toLowerCase().trim();
            
            if (searchTerm === '') {
                instance.filteredRows = [...instance.originalRows];
            } else {
                instance.filteredRows = instance.originalRows.filter(row => {
                    return row.textContent.toLowerCase().includes(searchTerm);
                });
            }

            instance.options.currentPage = 1;
            this.render(instance);
            this.updatePagination(instance);
        }, 300));
    },

    initSort(instance) {
        const headers = instance.table.querySelectorAll('th[data-sortable]');
        headers.forEach((header, index) => {
            header.style.cursor = 'pointer';
            header.classList.add('select-none', 'hover:bg-gray-100', 'dark:hover:bg-gray-700', 'transition-colors');
            
            // Add sort indicator
            const indicator = document.createElement('span');
            indicator.className = 'ml-1 text-gray-400 sort-indicator';
            indicator.innerHTML = `<svg class="w-4 h-4 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"/></svg>`;
            header.appendChild(indicator);

            header.addEventListener('click', () => this.sortTable(instance, header, index));
        });
    },

    sortTable(instance, header, columnIndex) {
        const isAsc = header.dataset.sortDir !== 'asc';
        
        // Reset all headers
        instance.table.querySelectorAll('th[data-sortable]').forEach(h => {
            h.dataset.sortDir = '';
            h.querySelector('.sort-indicator')?.classList.remove('text-primary-500');
        });

        header.dataset.sortDir = isAsc ? 'asc' : 'desc';
        header.querySelector('.sort-indicator')?.classList.add('text-primary-500');

        instance.filteredRows.sort((a, b) => {
            const aVal = a.children[columnIndex]?.textContent.trim() || '';
            const bVal = b.children[columnIndex]?.textContent.trim() || '';
            
            // Try numeric sort first
            const aNum = parseFloat(aVal.replace(/[^\d.-]/g, ''));
            const bNum = parseFloat(bVal.replace(/[^\d.-]/g, ''));
            
            if (!isNaN(aNum) && !isNaN(bNum)) {
                return isAsc ? aNum - bNum : bNum - aNum;
            }
            
            return isAsc ? aVal.localeCompare(bVal, 'fr') : bVal.localeCompare(aVal, 'fr');
        });

        this.render(instance);
    },

    initPagination(instance) {
        this.render(instance);
        this.updatePagination(instance);
    },

    render(instance) {
        const { perPage, currentPage } = instance.options;
        const start = (currentPage - 1) * perPage;
        const end = start + perPage;
        const tbody = instance.table.querySelector('tbody');

        // Hide all rows
        instance.originalRows.forEach(row => row.style.display = 'none');

        // Show only current page rows
        instance.filteredRows.slice(start, end).forEach(row => {
            row.style.display = '';
            tbody.appendChild(row);
        });
    },

    updatePagination(instance) {
        const paginationContainer = document.querySelector(`[data-table-pagination="${instance.table.id}"]`);
        if (!paginationContainer) return;

        const { perPage, currentPage } = instance.options;
        const totalRows = instance.filteredRows.length;
        const totalPages = Math.ceil(totalRows / perPage);
        const start = (currentPage - 1) * perPage + 1;
        const end = Math.min(currentPage * perPage, totalRows);

        paginationContainer.innerHTML = `
            <span class="text-sm text-gray-500 dark:text-gray-400">
                ${totalRows > 0 ? `${start}-${end} sur ${totalRows}` : 'Aucun résultat'}
            </span>
            <div class="flex items-center gap-1">
                <button 
                    class="p-2 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-500 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
                    ${currentPage === 1 ? 'disabled' : ''}
                    data-page="prev"
                >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
                </button>
                ${this.generatePageButtons(currentPage, totalPages)}
                <button 
                    class="p-2 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-500 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
                    ${currentPage === totalPages || totalPages === 0 ? 'disabled' : ''}
                    data-page="next"
                >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
                </button>
            </div>
        `;

        // Add event listeners
        paginationContainer.querySelectorAll('[data-page]').forEach(btn => {
            btn.addEventListener('click', () => {
                const page = btn.dataset.page;
                if (page === 'prev') {
                    instance.options.currentPage = Math.max(1, currentPage - 1);
                } else if (page === 'next') {
                    instance.options.currentPage = Math.min(totalPages, currentPage + 1);
                } else {
                    instance.options.currentPage = parseInt(page);
                }
                this.render(instance);
                this.updatePagination(instance);
            });
        });
    },

    generatePageButtons(current, total) {
        if (total <= 1) return '';
        
        let buttons = [];
        const maxVisible = 5;
        let start = Math.max(1, current - Math.floor(maxVisible / 2));
        let end = Math.min(total, start + maxVisible - 1);
        
        if (end - start < maxVisible - 1) {
            start = Math.max(1, end - maxVisible + 1);
        }

        for (let i = start; i <= end; i++) {
            const isActive = i === current;
            buttons.push(`
                <button 
                    class="w-9 h-9 rounded-lg text-sm font-medium ${isActive 
                        ? 'text-white bg-primary-500' 
                        : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'}"
                    data-page="${i}"
                >
                    ${i}
                </button>
            `);
        }

        return buttons.join('');
    },

    refresh(tableId) {
        const instance = this.instances.get(tableId);
        if (instance) {
            instance.originalRows = Array.from(instance.table.querySelectorAll('tbody tr'));
            instance.filteredRows = [...instance.originalRows];
            instance.options.currentPage = 1;
            this.render(instance);
            this.updatePagination(instance);
        }
    }
};

// ============================================
// MODAL - Modal Dialog Management
// ============================================
const Modal = {
    stack: [],

    open(modalId) {
        const modal = document.getElementById(modalId);
        if (!modal) return;

        modal.classList.remove('hidden');
        modal.setAttribute('aria-hidden', 'false');
        document.body.style.overflow = 'hidden';
        this.stack.push(modalId);

        // Focus first focusable element
        setTimeout(() => {
            const focusable = modal.querySelector('input, button, select, textarea, [tabindex]:not([tabindex="-1"])');
            focusable?.focus();
        }, 100);

        // Dispatch event
        modal.dispatchEvent(new CustomEvent('modal:opened', { detail: { modalId } }));
    },

    close(modalId) {
        const modal = document.getElementById(modalId);
        if (!modal) return;

        modal.classList.add('hidden');
        modal.setAttribute('aria-hidden', 'true');
        
        this.stack = this.stack.filter(id => id !== modalId);
        
        if (this.stack.length === 0) {
            document.body.style.overflow = '';
        }

        // Dispatch event
        modal.dispatchEvent(new CustomEvent('modal:closed', { detail: { modalId } }));
    },

    closeAll() {
        this.stack.forEach(modalId => this.close(modalId));
    },

    init() {
        // Close modal on backdrop click
        document.addEventListener('click', (e) => {
            if (e.target.matches('[data-modal-backdrop]')) {
                const modalId = e.target.closest('[data-modal]')?.id;
                if (modalId) this.close(modalId);
            }
        });

        // Close modal on close button click
        document.addEventListener('click', (e) => {
            const closeBtn = e.target.closest('[data-modal-close]');
            if (closeBtn) {
                const modalId = closeBtn.dataset.modalClose || closeBtn.closest('[data-modal]')?.id;
                if (modalId) this.close(modalId);
            }
        });

        // Close modal on escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.stack.length > 0) {
                this.close(this.stack[this.stack.length - 1]);
            }
        });

        // Open modal triggers
        document.addEventListener('click', (e) => {
            const trigger = e.target.closest('[data-modal-open]');
            if (trigger) {
                e.preventDefault();
                this.open(trigger.dataset.modalOpen);
            }
        });
    }
};

// ============================================
// TOAST - Notification System
// ============================================
const Toast = {
    container: null,
    queue: [],
    maxVisible: 5,

    init() {
        if (this.container) return;
        
        this.container = document.createElement('div');
        this.container.id = 'toast-container';
        this.container.className = 'fixed bottom-4 right-4 z-[9999] space-y-2 pointer-events-none';
        this.container.setAttribute('aria-live', 'polite');
        document.body.appendChild(this.container);
    },

    show(message, type = 'info', options = {}) {
        if (!this.container) this.init();

        const config = {
            duration: 5000,
            dismissible: true,
            icon: true,
            ...options
        };

        const colors = {
            success: 'bg-green-500',
            error: 'bg-red-500',
            warning: 'bg-amber-500',
            info: 'bg-blue-500'
        };

        const icons = {
            success: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>`,
            error: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>`,
            warning: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>`,
            info: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>`
        };

        const id = Utils.uniqueId('toast');
        const toast = document.createElement('div');
        toast.id = id;
        toast.className = `flex items-center gap-3 px-4 py-3 rounded-xl text-white ${colors[type]} shadow-lg transform translate-x-full transition-all duration-300 pointer-events-auto`;
        toast.setAttribute('role', 'alert');
        
        toast.innerHTML = `
            ${config.icon ? `<span class="flex-shrink-0">${icons[type]}</span>` : ''}
            <span class="text-sm font-medium flex-1">${Utils.escapeHtml(message)}</span>
            ${config.dismissible ? `
                <button class="flex-shrink-0 p-1 hover:bg-white/20 rounded-lg transition-colors" onclick="Toast.dismiss('${id}')">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
                </button>
            ` : ''}
        `;

        this.container.appendChild(toast);

        // Animate in
        requestAnimationFrame(() => {
            toast.classList.remove('translate-x-full');
        });

        // Auto dismiss
        if (config.duration > 0) {
            setTimeout(() => this.dismiss(id), config.duration);
        }

        return id;
    },

    dismiss(id) {
        const toast = document.getElementById(id);
        if (!toast) return;

        toast.classList.add('translate-x-full', 'opacity-0');
        setTimeout(() => toast.remove(), 300);
    },

    success(message, options = {}) {
        return this.show(message, 'success', options);
    },

    error(message, options = {}) {
        return this.show(message, 'error', options);
    },

    warning(message, options = {}) {
        return this.show(message, 'warning', options);
    },

    info(message, options = {}) {
        return this.show(message, 'info', options);
    },

    clear() {
        if (this.container) {
            this.container.innerHTML = '';
        }
    }
};

// ============================================
// FORM VALIDATOR - Client-side Validation
// ============================================
const FormValidator = {
    rules: {
        required: (value) => value.trim() !== '' || 'Ce champ est requis',
        email: (value) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value) || 'Email invalide',
        minLength: (min) => (value) => value.length >= min || `Minimum ${min} caractères`,
        maxLength: (max) => (value) => value.length <= max || `Maximum ${max} caractères`,
        pattern: (regex, msg) => (value) => regex.test(value) || msg,
        match: (fieldId, msg) => (value) => {
            const field = document.getElementById(fieldId);
            return field && value === field.value || msg || 'Les champs ne correspondent pas';
        },
        phone: (value) => /^[\d\s\-+()]+$/.test(value) || 'Numéro de téléphone invalide',
        url: (value) => {
            try { new URL(value); return true; } 
            catch { return 'URL invalide'; }
        },
        number: (value) => !isNaN(parseFloat(value)) || 'Nombre invalide',
        min: (min) => (value) => parseFloat(value) >= min || `Minimum ${min}`,
        max: (max) => (value) => parseFloat(value) <= max || `Maximum ${max}`
    },

    validate(form, customRules = {}) {
        const inputs = form.querySelectorAll('[data-validate]');
        let isValid = true;
        const errors = {};

        inputs.forEach(input => {
            const rules = input.dataset.validate.split('|');
            const value = input.value;
            const fieldName = input.name || input.id;

            for (const rule of rules) {
                let ruleName = rule;
                let ruleParam = null;

                if (rule.includes(':')) {
                    [ruleName, ruleParam] = rule.split(':');
                }

                const validator = customRules[ruleName] || this.rules[ruleName];
                if (!validator) continue;

                const validatorFn = typeof validator === 'function' && ruleParam 
                    ? validator(ruleParam) 
                    : validator;

                const result = validatorFn(value);
                
                if (result !== true) {
                    isValid = false;
                    errors[fieldName] = result;
                    this.showError(input, result);
                    break;
                } else {
                    this.clearError(input);
                }
            }
        });

        return { isValid, errors };
    },

    showError(input, message) {
        this.clearError(input);
        
        input.classList.add('border-red-500', 'focus:border-red-500', 'focus:ring-red-500');
        input.classList.remove('border-gray-300', 'focus:border-primary-500', 'focus:ring-primary-500');
        
        const error = document.createElement('p');
        error.className = 'text-red-500 text-xs mt-1 error-message';
        error.textContent = message;
        input.parentNode.appendChild(error);
    },

    clearError(input) {
        input.classList.remove('border-red-500', 'focus:border-red-500', 'focus:ring-red-500');
        input.classList.add('border-gray-300', 'focus:border-primary-500', 'focus:ring-primary-500');
        
        const error = input.parentNode.querySelector('.error-message');
        if (error) error.remove();
    },

    clearAll(form) {
        form.querySelectorAll('[data-validate]').forEach(input => this.clearError(input));
    }
};

// ============================================
// DROPDOWN - Dropdown Menu Management
// ============================================
const Dropdown = {
    activeDropdown: null,

    init() {
        // Toggle dropdown on trigger click
        document.addEventListener('click', (e) => {
            const trigger = e.target.closest('[data-dropdown-toggle]');
            
            if (trigger) {
                e.preventDefault();
                e.stopPropagation();
                const dropdownId = trigger.dataset.dropdownToggle;
                this.toggle(dropdownId);
                return;
            }

            // Close if clicking outside
            if (this.activeDropdown && !e.target.closest('[data-dropdown]')) {
                this.close(this.activeDropdown);
            }
        });

        // Close on escape
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.activeDropdown) {
                this.close(this.activeDropdown);
            }
        });
    },

    toggle(dropdownId) {
        const dropdown = document.getElementById(dropdownId);
        if (!dropdown) return;

        if (dropdown.classList.contains('hidden')) {
            this.open(dropdownId);
        } else {
            this.close(dropdownId);
        }
    },

    open(dropdownId) {
        // Close any open dropdown first
        if (this.activeDropdown && this.activeDropdown !== dropdownId) {
            this.close(this.activeDropdown);
        }

        const dropdown = document.getElementById(dropdownId);
        if (!dropdown) return;

        dropdown.classList.remove('hidden');
        this.activeDropdown = dropdownId;
    },

    close(dropdownId) {
        const dropdown = document.getElementById(dropdownId);
        if (!dropdown) return;

        dropdown.classList.add('hidden');
        if (this.activeDropdown === dropdownId) {
            this.activeDropdown = null;
        }
    }
};

// ============================================
// THEME - Dark Mode Management
// ============================================
const Theme = {
    init() {
        // Check for saved preference or system preference
        const savedTheme = localStorage.getItem('theme');
        const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        
        if (savedTheme === 'dark' || (!savedTheme && systemDark)) {
            document.documentElement.classList.add('dark');
        }

        // Listen for system theme changes
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            if (!localStorage.getItem('theme')) {
                document.documentElement.classList.toggle('dark', e.matches);
            }
        });
    },

    toggle() {
        const isDark = document.documentElement.classList.toggle('dark');
        localStorage.setItem('theme', isDark ? 'dark' : 'light');
        return isDark;
    },

    set(theme) {
        if (theme === 'dark') {
            document.documentElement.classList.add('dark');
        } else if (theme === 'light') {
            document.documentElement.classList.remove('dark');
        } else {
            localStorage.removeItem('theme');
            const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            document.documentElement.classList.toggle('dark', systemDark);
        }
        if (theme !== 'system') {
            localStorage.setItem('theme', theme);
        }
    },

    get() {
        return document.documentElement.classList.contains('dark') ? 'dark' : 'light';
    }
};

// ============================================
// SIDEBAR - Sidebar Management
// ============================================
const Sidebar = {
    isOpen: true,
    isMobile: false,

    init() {
        this.isMobile = window.innerWidth < 1024;
        this.isOpen = !this.isMobile;

        window.addEventListener('resize', Utils.debounce(() => {
            const wasMobile = this.isMobile;
            this.isMobile = window.innerWidth < 1024;
            
            if (wasMobile !== this.isMobile) {
                this.isOpen = !this.isMobile;
                this.update();
            }
        }, 100));
    },

    toggle() {
        this.isOpen = !this.isOpen;
        this.update();
    },

    open() {
        this.isOpen = true;
        this.update();
    },

    close() {
        this.isOpen = false;
        this.update();
    },

    update() {
        const sidebar = document.getElementById('sidebar');
        const overlay = document.getElementById('sidebar-overlay');
        const mainContent = document.getElementById('main-content');

        if (!sidebar) return;

        if (this.isMobile) {
            sidebar.classList.toggle('-translate-x-full', !this.isOpen);
            overlay?.classList.toggle('hidden', !this.isOpen);
        } else {
            sidebar.classList.remove('-translate-x-full');
            overlay?.classList.add('hidden');
            mainContent?.classList.toggle('lg:pl-64', this.isOpen);
            mainContent?.classList.toggle('lg:pl-0', !this.isOpen);
        }
    }
};

// ============================================
// HTMX INTEGRATION
// ============================================
const HTMXIntegration = {
    init() {
        // Show loading state on HTMX requests
        document.body.addEventListener('htmx:beforeRequest', (e) => {
            const target = e.detail.target;
            if (target) {
                target.classList.add('opacity-50', 'pointer-events-none');
            }
        });

        // Remove loading state after HTMX requests
        document.body.addEventListener('htmx:afterRequest', (e) => {
            const target = e.detail.target;
            if (target) {
                target.classList.remove('opacity-50', 'pointer-events-none');
            }
        });

        // Handle HTMX errors
        document.body.addEventListener('htmx:responseError', (e) => {
            Toast.error('Une erreur est survenue. Veuillez réessayer.');
        });

        // Reinitialize components after HTMX swap
        document.body.addEventListener('htmx:afterSwap', (e) => {
            // Reinit tables
            e.detail.target.querySelectorAll('[data-table]').forEach(table => {
                DataTable.init(table.id);
            });
        });

        // Handle flash messages from HTMX responses
        document.body.addEventListener('htmx:afterRequest', (e) => {
            const flashHeader = e.detail.xhr?.getResponseHeader('X-Flash-Message');
            const flashType = e.detail.xhr?.getResponseHeader('X-Flash-Type') || 'info';
            
            if (flashHeader) {
                Toast.show(flashHeader, flashType);
            }
        });
    }
};

// ============================================
// BULK ACTIONS - Mass Selection
// ============================================
const BulkActions = {
    init(containerId) {
        const container = document.getElementById(containerId);
        if (!container) return;

        const selectAll = container.querySelector('[data-select-all]');
        const checkboxes = container.querySelectorAll('[data-select-item]');
        const actionsBar = container.querySelector('[data-bulk-actions]');
        const countDisplay = container.querySelector('[data-selected-count]');

        if (!selectAll) return;

        selectAll.addEventListener('change', () => {
            checkboxes.forEach(cb => cb.checked = selectAll.checked);
            this.updateUI(container);
        });

        checkboxes.forEach(cb => {
            cb.addEventListener('change', () => {
                selectAll.checked = Array.from(checkboxes).every(c => c.checked);
                selectAll.indeterminate = !selectAll.checked && Array.from(checkboxes).some(c => c.checked);
                this.updateUI(container);
            });
        });
    },

    updateUI(container) {
        const checkboxes = container.querySelectorAll('[data-select-item]');
        const actionsBar = container.querySelector('[data-bulk-actions]');
        const countDisplay = container.querySelector('[data-selected-count]');
        
        const selected = Array.from(checkboxes).filter(cb => cb.checked);
        
        if (actionsBar) {
            actionsBar.classList.toggle('hidden', selected.length === 0);
        }
        
        if (countDisplay) {
            countDisplay.textContent = selected.length;
        }
    },

    getSelected(containerId) {
        const container = document.getElementById(containerId);
        if (!container) return [];
        
        return Array.from(container.querySelectorAll('[data-select-item]:checked'))
            .map(cb => cb.value || cb.dataset.id);
    },

    clearSelection(containerId) {
        const container = document.getElementById(containerId);
        if (!container) return;
        
        container.querySelectorAll('[data-select-all], [data-select-item]').forEach(cb => {
            cb.checked = false;
            cb.indeterminate = false;
        });
        this.updateUI(container);
    }
};

// ============================================
// INITIALIZATION
// ============================================
document.addEventListener('DOMContentLoaded', () => {
    // Initialize all modules
    Theme.init();
    Modal.init();
    Toast.init();
    Dropdown.init();
    Sidebar.init();
    HTMXIntegration.init();

    // Initialize tables
    document.querySelectorAll('[data-table]').forEach(table => {
        DataTable.init(table.id);
    });

    // Initialize bulk actions
    document.querySelectorAll('[data-bulk-container]').forEach(container => {
        BulkActions.init(container.id);
    });
});

// ============================================
// GLOBAL EXPORTS
// ============================================
window.SublimeGo = {
    Utils,
    DataTable,
    Modal,
    Toast,
    FormValidator,
    Dropdown,
    Theme,
    Sidebar,
    BulkActions
};

// Shortcuts
window.Utils = Utils;
window.DataTable = DataTable;
window.Modal = Modal;
window.Toast = Toast;
window.FormValidator = FormValidator;
window.Dropdown = Dropdown;
window.Theme = Theme;
window.Sidebar = Sidebar;
window.BulkActions = BulkActions;
