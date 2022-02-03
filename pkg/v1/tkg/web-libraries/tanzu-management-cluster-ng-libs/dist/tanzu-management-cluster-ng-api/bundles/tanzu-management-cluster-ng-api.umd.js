(function (global, factory) {
    typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports, require('@angular/core'), require('@angular/common/http'), require('rxjs'), require('rxjs/operators')) :
    typeof define === 'function' && define.amd ? define('tanzu-management-cluster-ng-api', ['exports', '@angular/core', '@angular/common/http', 'rxjs', 'rxjs/operators'], factory) :
    (global = typeof globalThis !== 'undefined' ? globalThis : global || self, factory(global["tanzu-management-cluster-ng-api"] = {}, global.ng.core, global.ng.common.http, global.rxjs, global.rxjs.operators));
})(this, (function (exports, i0, i1, rxjs, operators) { 'use strict';

    function _interopNamespace(e) {
        if (e && e.__esModule) return e;
        var n = Object.create(null);
        if (e) {
            Object.keys(e).forEach(function (k) {
                if (k !== 'default') {
                    var d = Object.getOwnPropertyDescriptor(e, k);
                    Object.defineProperty(n, k, d.get ? d : {
                        enumerable: true,
                        get: function () { return e[k]; }
                    });
                }
            });
        }
        n["default"] = e;
        return Object.freeze(n);
    }

    var i0__namespace = /*#__PURE__*/_interopNamespace(i0);
    var i1__namespace = /*#__PURE__*/_interopNamespace(i1);

    /*! *****************************************************************************
    Copyright (c) Microsoft Corporation.

    Permission to use, copy, modify, and/or distribute this software for any
    purpose with or without fee is hereby granted.

    THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
    REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
    AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
    INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
    LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
    OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
    PERFORMANCE OF THIS SOFTWARE.
    ***************************************************************************** */
    /* global Reflect, Promise */
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b)
                if (Object.prototype.hasOwnProperty.call(b, p))
                    d[p] = b[p]; };
        return extendStatics(d, b);
    };
    function __extends(d, b) {
        if (typeof b !== "function" && b !== null)
            throw new TypeError("Class extends value " + String(b) + " is not a constructor or null");
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    }
    var __assign = function () {
        __assign = Object.assign || function __assign(t) {
            for (var s, i = 1, n = arguments.length; i < n; i++) {
                s = arguments[i];
                for (var p in s)
                    if (Object.prototype.hasOwnProperty.call(s, p))
                        t[p] = s[p];
            }
            return t;
        };
        return __assign.apply(this, arguments);
    };
    function __rest(s, e) {
        var t = {};
        for (var p in s)
            if (Object.prototype.hasOwnProperty.call(s, p) && e.indexOf(p) < 0)
                t[p] = s[p];
        if (s != null && typeof Object.getOwnPropertySymbols === "function")
            for (var i = 0, p = Object.getOwnPropertySymbols(s); i < p.length; i++) {
                if (e.indexOf(p[i]) < 0 && Object.prototype.propertyIsEnumerable.call(s, p[i]))
                    t[p[i]] = s[p[i]];
            }
        return t;
    }
    function __decorate(decorators, target, key, desc) {
        var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
        if (typeof Reflect === "object" && typeof Reflect.decorate === "function")
            r = Reflect.decorate(decorators, target, key, desc);
        else
            for (var i = decorators.length - 1; i >= 0; i--)
                if (d = decorators[i])
                    r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
        return c > 3 && r && Object.defineProperty(target, key, r), r;
    }
    function __param(paramIndex, decorator) {
        return function (target, key) { decorator(target, key, paramIndex); };
    }
    function __metadata(metadataKey, metadataValue) {
        if (typeof Reflect === "object" && typeof Reflect.metadata === "function")
            return Reflect.metadata(metadataKey, metadataValue);
    }
    function __awaiter(thisArg, _arguments, P, generator) {
        function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
        return new (P || (P = Promise))(function (resolve, reject) {
            function fulfilled(value) { try {
                step(generator.next(value));
            }
            catch (e) {
                reject(e);
            } }
            function rejected(value) { try {
                step(generator["throw"](value));
            }
            catch (e) {
                reject(e);
            } }
            function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
            step((generator = generator.apply(thisArg, _arguments || [])).next());
        });
    }
    function __generator(thisArg, body) {
        var _ = { label: 0, sent: function () { if (t[0] & 1)
                throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g;
        return g = { next: verb(0), "throw": verb(1), "return": verb(2) }, typeof Symbol === "function" && (g[Symbol.iterator] = function () { return this; }), g;
        function verb(n) { return function (v) { return step([n, v]); }; }
        function step(op) {
            if (f)
                throw new TypeError("Generator is already executing.");
            while (_)
                try {
                    if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done)
                        return t;
                    if (y = 0, t)
                        op = [op[0] & 2, t.value];
                    switch (op[0]) {
                        case 0:
                        case 1:
                            t = op;
                            break;
                        case 4:
                            _.label++;
                            return { value: op[1], done: false };
                        case 5:
                            _.label++;
                            y = op[1];
                            op = [0];
                            continue;
                        case 7:
                            op = _.ops.pop();
                            _.trys.pop();
                            continue;
                        default:
                            if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) {
                                _ = 0;
                                continue;
                            }
                            if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) {
                                _.label = op[1];
                                break;
                            }
                            if (op[0] === 6 && _.label < t[1]) {
                                _.label = t[1];
                                t = op;
                                break;
                            }
                            if (t && _.label < t[2]) {
                                _.label = t[2];
                                _.ops.push(op);
                                break;
                            }
                            if (t[2])
                                _.ops.pop();
                            _.trys.pop();
                            continue;
                    }
                    op = body.call(thisArg, _);
                }
                catch (e) {
                    op = [6, e];
                    y = 0;
                }
                finally {
                    f = t = 0;
                }
            if (op[0] & 5)
                throw op[1];
            return { value: op[0] ? op[1] : void 0, done: true };
        }
    }
    var __createBinding = Object.create ? (function (o, m, k, k2) {
        if (k2 === undefined)
            k2 = k;
        Object.defineProperty(o, k2, { enumerable: true, get: function () { return m[k]; } });
    }) : (function (o, m, k, k2) {
        if (k2 === undefined)
            k2 = k;
        o[k2] = m[k];
    });
    function __exportStar(m, o) {
        for (var p in m)
            if (p !== "default" && !Object.prototype.hasOwnProperty.call(o, p))
                __createBinding(o, m, p);
    }
    function __values(o) {
        var s = typeof Symbol === "function" && Symbol.iterator, m = s && o[s], i = 0;
        if (m)
            return m.call(o);
        if (o && typeof o.length === "number")
            return {
                next: function () {
                    if (o && i >= o.length)
                        o = void 0;
                    return { value: o && o[i++], done: !o };
                }
            };
        throw new TypeError(s ? "Object is not iterable." : "Symbol.iterator is not defined.");
    }
    function __read(o, n) {
        var m = typeof Symbol === "function" && o[Symbol.iterator];
        if (!m)
            return o;
        var i = m.call(o), r, ar = [], e;
        try {
            while ((n === void 0 || n-- > 0) && !(r = i.next()).done)
                ar.push(r.value);
        }
        catch (error) {
            e = { error: error };
        }
        finally {
            try {
                if (r && !r.done && (m = i["return"]))
                    m.call(i);
            }
            finally {
                if (e)
                    throw e.error;
            }
        }
        return ar;
    }
    /** @deprecated */
    function __spread() {
        for (var ar = [], i = 0; i < arguments.length; i++)
            ar = ar.concat(__read(arguments[i]));
        return ar;
    }
    /** @deprecated */
    function __spreadArrays() {
        for (var s = 0, i = 0, il = arguments.length; i < il; i++)
            s += arguments[i].length;
        for (var r = Array(s), k = 0, i = 0; i < il; i++)
            for (var a = arguments[i], j = 0, jl = a.length; j < jl; j++, k++)
                r[k] = a[j];
        return r;
    }
    function __spreadArray(to, from, pack) {
        if (pack || arguments.length === 2)
            for (var i = 0, l = from.length, ar; i < l; i++) {
                if (ar || !(i in from)) {
                    if (!ar)
                        ar = Array.prototype.slice.call(from, 0, i);
                    ar[i] = from[i];
                }
            }
        return to.concat(ar || Array.prototype.slice.call(from));
    }
    function __await(v) {
        return this instanceof __await ? (this.v = v, this) : new __await(v);
    }
    function __asyncGenerator(thisArg, _arguments, generator) {
        if (!Symbol.asyncIterator)
            throw new TypeError("Symbol.asyncIterator is not defined.");
        var g = generator.apply(thisArg, _arguments || []), i, q = [];
        return i = {}, verb("next"), verb("throw"), verb("return"), i[Symbol.asyncIterator] = function () { return this; }, i;
        function verb(n) { if (g[n])
            i[n] = function (v) { return new Promise(function (a, b) { q.push([n, v, a, b]) > 1 || resume(n, v); }); }; }
        function resume(n, v) { try {
            step(g[n](v));
        }
        catch (e) {
            settle(q[0][3], e);
        } }
        function step(r) { r.value instanceof __await ? Promise.resolve(r.value.v).then(fulfill, reject) : settle(q[0][2], r); }
        function fulfill(value) { resume("next", value); }
        function reject(value) { resume("throw", value); }
        function settle(f, v) { if (f(v), q.shift(), q.length)
            resume(q[0][0], q[0][1]); }
    }
    function __asyncDelegator(o) {
        var i, p;
        return i = {}, verb("next"), verb("throw", function (e) { throw e; }), verb("return"), i[Symbol.iterator] = function () { return this; }, i;
        function verb(n, f) { i[n] = o[n] ? function (v) { return (p = !p) ? { value: __await(o[n](v)), done: n === "return" } : f ? f(v) : v; } : f; }
    }
    function __asyncValues(o) {
        if (!Symbol.asyncIterator)
            throw new TypeError("Symbol.asyncIterator is not defined.");
        var m = o[Symbol.asyncIterator], i;
        return m ? m.call(o) : (o = typeof __values === "function" ? __values(o) : o[Symbol.iterator](), i = {}, verb("next"), verb("throw"), verb("return"), i[Symbol.asyncIterator] = function () { return this; }, i);
        function verb(n) { i[n] = o[n] && function (v) { return new Promise(function (resolve, reject) { v = o[n](v), settle(resolve, reject, v.done, v.value); }); }; }
        function settle(resolve, reject, d, v) { Promise.resolve(v).then(function (v) { resolve({ value: v, done: d }); }, reject); }
    }
    function __makeTemplateObject(cooked, raw) {
        if (Object.defineProperty) {
            Object.defineProperty(cooked, "raw", { value: raw });
        }
        else {
            cooked.raw = raw;
        }
        return cooked;
    }
    ;
    var __setModuleDefault = Object.create ? (function (o, v) {
        Object.defineProperty(o, "default", { enumerable: true, value: v });
    }) : function (o, v) {
        o["default"] = v;
    };
    function __importStar(mod) {
        if (mod && mod.__esModule)
            return mod;
        var result = {};
        if (mod != null)
            for (var k in mod)
                if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k))
                    __createBinding(result, mod, k);
        __setModuleDefault(result, mod);
        return result;
    }
    function __importDefault(mod) {
        return (mod && mod.__esModule) ? mod : { default: mod };
    }
    function __classPrivateFieldGet(receiver, state, kind, f) {
        if (kind === "a" && !f)
            throw new TypeError("Private accessor was defined without a getter");
        if (typeof state === "function" ? receiver !== state || !f : !state.has(receiver))
            throw new TypeError("Cannot read private member from an object whose class did not declare it");
        return kind === "m" ? f : kind === "a" ? f.call(receiver) : f ? f.value : state.get(receiver);
    }
    function __classPrivateFieldSet(receiver, state, value, kind, f) {
        if (kind === "m")
            throw new TypeError("Private method is not writable");
        if (kind === "a" && !f)
            throw new TypeError("Private accessor was defined without a setter");
        if (typeof state === "function" ? receiver !== state || !f : !state.has(receiver))
            throw new TypeError("Cannot write private member to an object whose class did not declare it");
        return (kind === "a" ? f.call(receiver, value) : f ? f.value = value : state.set(receiver, value)), value;
    }

    /* tslint:disable */
    var USE_DOMAIN = new i0.InjectionToken('APIClient_USE_DOMAIN');
    var USE_HTTP_OPTIONS = new i0.InjectionToken('APIClient_USE_HTTP_OPTIONS');
    /**
     * Created with https://github.com/flowup/api-client-generator
     */
    var APIClient = /** @class */ (function () {
        function APIClient(http, domain, options) {
            this.http = http;
            this.domain = "//" + window.location.hostname + (window.location.port ? ':' + window.location.port : '');
            if (domain != null) {
                this.domain = domain;
            }
            this.options = Object.assign(Object.assign({ headers: new i1.HttpHeaders(options && options.headers ? options.headers : {}), params: new i1.HttpParams(options && options.params ? options.params : {}) }, (options && options.reportProgress ? { reportProgress: options.reportProgress } : {})), (options && options.withCredentials ? { withCredentials: options.withCredentials } : {}));
        }
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getUI = function (requestHttpOptions) {
            var path = "/";
            var options = Object.assign(Object.assign(Object.assign({}, this.options), requestHttpOptions), { responseType: 'blob' });
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getUIFile = function (args, requestHttpOptions) {
            var path = "/" + args.filename;
            var options = Object.assign(Object.assign(Object.assign({}, this.options), requestHttpOptions), { responseType: 'blob' });
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getFeatureFlags = function (requestHttpOptions) {
            var path = "/api/features";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getTanzuEdition = function (requestHttpOptions) {
            var path = "/api/edition";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 201 ] HTTP response code.
         */
        APIClient.prototype.verifyAccount = function (args, requestHttpOptions) {
            var path = "/api/avi";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.credentials));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.verifyLdapConnect = function (args, requestHttpOptions) {
            var path = "/api/ldap/connect";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.credentials));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.verifyLdapBind = function (requestHttpOptions) {
            var path = "/api/ldap/bind";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.verifyLdapUserSearch = function (requestHttpOptions) {
            var path = "/api/ldap/users/search";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.verifyLdapGroupSearch = function (requestHttpOptions) {
            var path = "/api/ldap/groups/search";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options);
        };
        /**
         * Response generated for [ 201 ] HTTP response code.
         */
        APIClient.prototype.verifyLdapCloseConnection = function (requestHttpOptions) {
            var path = "/api/ldap/disconnect";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAviClouds = function (requestHttpOptions) {
            var path = "/api/avi/clouds";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAviServiceEngineGroups = function (requestHttpOptions) {
            var path = "/api/avi/serviceenginegroups";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAviVipNetworks = function (requestHttpOptions) {
            var path = "/api/avi/vipnetworks";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getProvider = function (requestHttpOptions) {
            var path = "/api/providers";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getVsphereThumbprint = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere/thumbprint";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('host' in args) {
                options.params = options.params.set('host', String(args.host));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 201 ] HTTP response code.
         */
        APIClient.prototype.setVSphereEndpoint = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.credentials));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getVSphereDatacenters = function (requestHttpOptions) {
            var path = "/api/providers/vsphere/datacenters";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getVSphereDatastores = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere/datastores";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('dc' in args) {
                options.params = options.params.set('dc', String(args.dc));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getVSphereFolders = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere/folders";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('dc' in args) {
                options.params = options.params.set('dc', String(args.dc));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getVSphereComputeResources = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere/compute";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('dc' in args) {
                options.params = options.params.set('dc', String(args.dc));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getVSphereResourcePools = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere/resourcepools";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('dc' in args) {
                options.params = options.params.set('dc', String(args.dc));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getVSphereNetworks = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere/networks";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('dc' in args) {
                options.params = options.params.set('dc', String(args.dc));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getVSphereNodeTypes = function (requestHttpOptions) {
            var path = "/api/providers/vsphere/nodetypes";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getVSphereOSImages = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere/osimages";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('dc' in args) {
                options.params = options.params.set('dc', String(args.dc));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.exportTKGConfigForVsphere = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere/config/export";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.applyTKGConfigForVsphere = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere/tkgconfig";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.importTKGConfigForVsphere = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere/config/import";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.createVSphereRegionalCluster = function (args, requestHttpOptions) {
            var path = "/api/providers/vsphere/create";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 201 ] HTTP response code.
         */
        APIClient.prototype.setAWSEndpoint = function (args, requestHttpOptions) {
            var path = "/api/providers/aws";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.accountParams));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getVPCs = function (requestHttpOptions) {
            var path = "/api/providers/aws/vpc";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAWSNodeTypes = function (args, requestHttpOptions) {
            var path = "/api/providers/aws/nodetypes";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('az' in args) {
                options.params = options.params.set('az', String(args.az));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAWSRegions = function (requestHttpOptions) {
            var path = "/api/providers/aws/regions";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAWSOSImages = function (args, requestHttpOptions) {
            var path = "/api/providers/aws/osimages";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('region' in args) {
                options.params = options.params.set('region', String(args.region));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAWSCredentialProfiles = function (requestHttpOptions) {
            var path = "/api/providers/aws/profiles";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAWSAvailabilityZones = function (requestHttpOptions) {
            var path = "/api/providers/aws/AvailabilityZones";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAWSSubnets = function (args, requestHttpOptions) {
            var path = "/api/providers/aws/subnets";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('vpcId' in args) {
                options.params = options.params.set('vpcId', String(args.vpcId));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.exportTKGConfigForAWS = function (args, requestHttpOptions) {
            var path = "/api/providers/aws/config/export";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.applyTKGConfigForAWS = function (args, requestHttpOptions) {
            var path = "/api/providers/aws/tkgconfig";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.createAWSRegionalCluster = function (args, requestHttpOptions) {
            var path = "/api/providers/aws/create";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.importTKGConfigForAWS = function (args, requestHttpOptions) {
            var path = "/api/providers/aws/config/import";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAzureEndpoint = function (requestHttpOptions) {
            var path = "/api/providers/azure";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 201 ] HTTP response code.
         */
        APIClient.prototype.setAzureEndpoint = function (args, requestHttpOptions) {
            var path = "/api/providers/azure";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.accountParams));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAzureResourceGroups = function (args, requestHttpOptions) {
            var path = "/api/providers/azure/resourcegroups";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('location' in args) {
                options.params = options.params.set('location', String(args.location));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 201 ] HTTP response code.
         */
        APIClient.prototype.createAzureResourceGroup = function (args, requestHttpOptions) {
            var path = "/api/providers/azure/resourcegroups";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAzureVnets = function (args, requestHttpOptions) {
            var path = "/api/providers/azure/resourcegroups/" + args.resourceGroupName + "/vnets";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            if ('location' in args) {
                options.params = options.params.set('location', String(args.location));
            }
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 201 ] HTTP response code.
         */
        APIClient.prototype.createAzureVirtualNetwork = function (args, requestHttpOptions) {
            var path = "/api/providers/azure/resourcegroups/" + args.resourceGroupName + "/vnets";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAzureOSImages = function (requestHttpOptions) {
            var path = "/api/providers/azure/osimages";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAzureRegions = function (requestHttpOptions) {
            var path = "/api/providers/azure/regions";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.getAzureInstanceTypes = function (args, requestHttpOptions) {
            var path = "/api/providers/azure/regions/" + args.location + "/instanceTypes";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.exportTKGConfigForAzure = function (args, requestHttpOptions) {
            var path = "/api/providers/azure/config/export";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.applyTKGConfigForAzure = function (args, requestHttpOptions) {
            var path = "/api/providers/azure/tkgconfig";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.createAzureRegionalCluster = function (args, requestHttpOptions) {
            var path = "/api/providers/azure/create";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.importTKGConfigForAzure = function (args, requestHttpOptions) {
            var path = "/api/providers/azure/config/import";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.checkIfDockerDaemonAvailable = function (requestHttpOptions) {
            var path = "/api/providers/docker/daemon";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('GET', path, options);
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.exportTKGConfigForDocker = function (args, requestHttpOptions) {
            var path = "/api/providers/docker/config/export";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.applyTKGConfigForDocker = function (args, requestHttpOptions) {
            var path = "/api/providers/docker/tkgconfig";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.createDockerRegionalCluster = function (args, requestHttpOptions) {
            var path = "/api/providers/docker/create";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        /**
         * Response generated for [ 200 ] HTTP response code.
         */
        APIClient.prototype.importTKGConfigForDocker = function (args, requestHttpOptions) {
            var path = "/api/providers/docker/config/import";
            var options = Object.assign(Object.assign({}, this.options), requestHttpOptions);
            return this.sendRequest('POST', path, options, JSON.stringify(args.params));
        };
        APIClient.prototype.sendRequest = function (method, path, options, body) {
            switch (method) {
                case 'DELETE':
                    return this.http.delete("" + this.domain + path, options);
                case 'GET':
                    return this.http.get("" + this.domain + path, options);
                case 'HEAD':
                    return this.http.head("" + this.domain + path, options);
                case 'OPTIONS':
                    return this.http.options("" + this.domain + path, options);
                case 'PATCH':
                    return this.http.patch("" + this.domain + path, body, options);
                case 'POST':
                    return this.http.post("" + this.domain + path, body, options);
                case 'PUT':
                    return this.http.put("" + this.domain + path, body, options);
                default:
                    console.error("Unsupported request: " + method);
                    return rxjs.throwError("Unsupported request: " + method);
            }
        };
        return APIClient;
    }());
    APIClient.ɵfac = i0__namespace.ɵɵngDeclareFactory({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0__namespace, type: APIClient, deps: [{ token: i1__namespace.HttpClient }, { token: USE_DOMAIN, optional: true }, { token: USE_HTTP_OPTIONS, optional: true }], target: i0__namespace.ɵɵFactoryTarget.Injectable });
    APIClient.ɵprov = i0__namespace.ɵɵngDeclareInjectable({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0__namespace, type: APIClient });
    i0__namespace.ɵɵngDeclareClassMetadata({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0__namespace, type: APIClient, decorators: [{
                type: i0.Injectable
            }], ctorParameters: function () {
            return [{ type: i1__namespace.HttpClient }, { type: undefined, decorators: [{
                            type: i0.Optional
                        }, {
                            type: i0.Inject,
                            args: [USE_DOMAIN]
                        }] }, { type: undefined, decorators: [{
                            type: i0.Optional
                        }, {
                            type: i0.Inject,
                            args: [USE_HTTP_OPTIONS]
                        }] }];
        } });

    /* tslint:disable */
    /* pre-prepared guards for build in complex types */
    function _isBlob(arg) {
        return arg != null && typeof arg.size === 'number' && typeof arg.type === 'string' && typeof arg.slice === 'function';
    }
    function isFile(arg) {
        return arg != null && typeof arg.lastModified === 'number' && typeof arg.name === 'string' && _isBlob(arg);
    }
    /* generated type guards */
    function isAviCloud(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // location?: string
            (typeof arg.location === 'undefined' || typeof arg.location === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // uuid?: string
            (typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string') &&
            true);
    }
    function isAviConfig(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // ca_cert?: string
            (typeof arg.ca_cert === 'undefined' || typeof arg.ca_cert === 'string') &&
            // cloud?: string
            (typeof arg.cloud === 'undefined' || typeof arg.cloud === 'string') &&
            // controller?: string
            (typeof arg.controller === 'undefined' || typeof arg.controller === 'string') &&
            // controlPlaneHaProvider?: boolean
            (typeof arg.controlPlaneHaProvider === 'undefined' || typeof arg.controlPlaneHaProvider === 'boolean') &&
            // labels?: { [key: string]: string }
            (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
            // managementClusterVipNetworkCidr?: string
            (typeof arg.managementClusterVipNetworkCidr === 'undefined' || typeof arg.managementClusterVipNetworkCidr === 'string') &&
            // managementClusterVipNetworkName?: string
            (typeof arg.managementClusterVipNetworkName === 'undefined' || typeof arg.managementClusterVipNetworkName === 'string') &&
            // network?: AviNetworkParams
            (typeof arg.network === 'undefined' || isAviNetworkParams(arg.network)) &&
            // password?: string
            (typeof arg.password === 'undefined' || typeof arg.password === 'string') &&
            // service_engine?: string
            (typeof arg.service_engine === 'undefined' || typeof arg.service_engine === 'string') &&
            // username?: string
            (typeof arg.username === 'undefined' || typeof arg.username === 'string') &&
            true);
    }
    function isAviControllerParams(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // CAData?: string
            (typeof arg.CAData === 'undefined' || typeof arg.CAData === 'string') &&
            // host?: string
            (typeof arg.host === 'undefined' || typeof arg.host === 'string') &&
            // password?: string
            (typeof arg.password === 'undefined' || typeof arg.password === 'string') &&
            // tenant?: string
            (typeof arg.tenant === 'undefined' || typeof arg.tenant === 'string') &&
            // username?: string
            (typeof arg.username === 'undefined' || typeof arg.username === 'string') &&
            true);
    }
    function isAviNetworkParams(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // cidr?: string
            (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            true);
    }
    function isAviServiceEngineGroup(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // location?: string
            (typeof arg.location === 'undefined' || typeof arg.location === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // uuid?: string
            (typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string') &&
            true);
    }
    function isAviSubnet(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // family?: string
            (typeof arg.family === 'undefined' || typeof arg.family === 'string') &&
            // subnet?: string
            (typeof arg.subnet === 'undefined' || typeof arg.subnet === 'string') &&
            true);
    }
    function isAviVipNetwork(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // cloud?: string
            (typeof arg.cloud === 'undefined' || typeof arg.cloud === 'string') &&
            // configedSubnets?: AviSubnet[]
            (typeof arg.configedSubnets === 'undefined' || (Array.isArray(arg.configedSubnets) && arg.configedSubnets.every(function (item) { return isAviSubnet(item); }))) &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // uuid?: string
            (typeof arg.uuid === 'undefined' || typeof arg.uuid === 'string') &&
            true);
    }
    function isAWSAccountParams(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // accessKeyID?: string
            (typeof arg.accessKeyID === 'undefined' || typeof arg.accessKeyID === 'string') &&
            // profileName?: string
            (typeof arg.profileName === 'undefined' || typeof arg.profileName === 'string') &&
            // region?: string
            (typeof arg.region === 'undefined' || typeof arg.region === 'string') &&
            // secretAccessKey?: string
            (typeof arg.secretAccessKey === 'undefined' || typeof arg.secretAccessKey === 'string') &&
            // sessionToken?: string
            (typeof arg.sessionToken === 'undefined' || typeof arg.sessionToken === 'string') &&
            true);
    }
    function isAWSAvailabilityZone(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // id?: string
            (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            true);
    }
    function isAWSNodeAz(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // privateSubnetID?: string
            (typeof arg.privateSubnetID === 'undefined' || typeof arg.privateSubnetID === 'string') &&
            // publicSubnetID?: string
            (typeof arg.publicSubnetID === 'undefined' || typeof arg.publicSubnetID === 'string') &&
            // workerNodeType?: string
            (typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string') &&
            true);
    }
    function isAWSRegionalClusterParams(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // annotations?: { [key: string]: string }
            (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
            // awsAccountParams?: AWSAccountParams
            (typeof arg.awsAccountParams === 'undefined' || isAWSAccountParams(arg.awsAccountParams)) &&
            // bastionHostEnabled?: boolean
            (typeof arg.bastionHostEnabled === 'undefined' || typeof arg.bastionHostEnabled === 'boolean') &&
            // ceipOptIn?: boolean
            (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
            // clusterName?: string
            (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
            // controlPlaneFlavor?: string
            (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
            // controlPlaneNodeType?: string
            (typeof arg.controlPlaneNodeType === 'undefined' || typeof arg.controlPlaneNodeType === 'string') &&
            // createCloudFormationStack?: boolean
            (typeof arg.createCloudFormationStack === 'undefined' || typeof arg.createCloudFormationStack === 'boolean') &&
            // enableAuditLogging?: boolean
            (typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean') &&
            // identityManagement?: IdentityManagementConfig
            (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
            // kubernetesVersion?: string
            (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
            // labels?: { [key: string]: string }
            (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
            // loadbalancerSchemeInternal?: boolean
            (typeof arg.loadbalancerSchemeInternal === 'undefined' || typeof arg.loadbalancerSchemeInternal === 'boolean') &&
            // machineHealthCheckEnabled?: boolean
            (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
            // networking?: TKGNetwork
            (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
            // numOfWorkerNode?: number
            (typeof arg.numOfWorkerNode === 'undefined' || typeof arg.numOfWorkerNode === 'number') &&
            // os?: AWSVirtualMachine
            (typeof arg.os === 'undefined' || isAWSVirtualMachine(arg.os)) &&
            // sshKeyName?: string
            (typeof arg.sshKeyName === 'undefined' || typeof arg.sshKeyName === 'string') &&
            // vpc?: AWSVpc
            (typeof arg.vpc === 'undefined' || isAWSVpc(arg.vpc)) &&
            // workerNodeType?: string
            (typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string') &&
            true);
    }
    function isAWSRoute(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // DestinationCidrBlock?: string
            (typeof arg.DestinationCidrBlock === 'undefined' || typeof arg.DestinationCidrBlock === 'string') &&
            // GatewayId?: string
            (typeof arg.GatewayId === 'undefined' || typeof arg.GatewayId === 'string') &&
            // State?: string
            (typeof arg.State === 'undefined' || typeof arg.State === 'string') &&
            true);
    }
    function isAWSRouteTable(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // id?: string
            (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
            // routes?: AWSRoute[]
            (typeof arg.routes === 'undefined' || (Array.isArray(arg.routes) && arg.routes.every(function (item) { return isAWSRoute(item); }))) &&
            // vpcId?: string
            (typeof arg.vpcId === 'undefined' || typeof arg.vpcId === 'string') &&
            true);
    }
    function isAWSSubnet(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // availabilityZoneId?: string
            (typeof arg.availabilityZoneId === 'undefined' || typeof arg.availabilityZoneId === 'string') &&
            // availabilityZoneName?: string
            (typeof arg.availabilityZoneName === 'undefined' || typeof arg.availabilityZoneName === 'string') &&
            // cidr?: string
            (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
            // id?: string
            (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
            // isPublic: boolean
            (typeof arg.isPublic === 'boolean') &&
            // state?: string
            (typeof arg.state === 'undefined' || typeof arg.state === 'string') &&
            // vpcId?: string
            (typeof arg.vpcId === 'undefined' || typeof arg.vpcId === 'string') &&
            true);
    }
    function isAWSVirtualMachine(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // osInfo?: OSInfo
            (typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo)) &&
            true);
    }
    function isAWSVpc(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // azs?: AWSNodeAz[]
            (typeof arg.azs === 'undefined' || (Array.isArray(arg.azs) && arg.azs.every(function (item) { return isAWSNodeAz(item); }))) &&
            // cidr?: string
            (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
            // vpcID?: string
            (typeof arg.vpcID === 'undefined' || typeof arg.vpcID === 'string') &&
            true);
    }
    function isAzureAccountParams(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // azureCloud?: string
            (typeof arg.azureCloud === 'undefined' || typeof arg.azureCloud === 'string') &&
            // clientId?: string
            (typeof arg.clientId === 'undefined' || typeof arg.clientId === 'string') &&
            // clientSecret?: string
            (typeof arg.clientSecret === 'undefined' || typeof arg.clientSecret === 'string') &&
            // subscriptionId?: string
            (typeof arg.subscriptionId === 'undefined' || typeof arg.subscriptionId === 'string') &&
            // tenantId?: string
            (typeof arg.tenantId === 'undefined' || typeof arg.tenantId === 'string') &&
            true);
    }
    function isAzureInstanceType(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // family?: string
            (typeof arg.family === 'undefined' || typeof arg.family === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // size?: string
            (typeof arg.size === 'undefined' || typeof arg.size === 'string') &&
            // tier?: string
            (typeof arg.tier === 'undefined' || typeof arg.tier === 'string') &&
            // zones?: string[]
            (typeof arg.zones === 'undefined' || (Array.isArray(arg.zones) && arg.zones.every(function (item) { return typeof item === 'string'; }))) &&
            true);
    }
    function isAzureLocation(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // displayName?: string
            (typeof arg.displayName === 'undefined' || typeof arg.displayName === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            true);
    }
    function isAzureRegionalClusterParams(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // annotations?: { [key: string]: string }
            (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
            // azureAccountParams?: AzureAccountParams
            (typeof arg.azureAccountParams === 'undefined' || isAzureAccountParams(arg.azureAccountParams)) &&
            // ceipOptIn?: boolean
            (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
            // clusterName?: string
            (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
            // controlPlaneFlavor?: string
            (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
            // controlPlaneMachineType?: string
            (typeof arg.controlPlaneMachineType === 'undefined' || typeof arg.controlPlaneMachineType === 'string') &&
            // controlPlaneSubnet?: string
            (typeof arg.controlPlaneSubnet === 'undefined' || typeof arg.controlPlaneSubnet === 'string') &&
            // controlPlaneSubnetCidr?: string
            (typeof arg.controlPlaneSubnetCidr === 'undefined' || typeof arg.controlPlaneSubnetCidr === 'string') &&
            // enableAuditLogging?: boolean
            (typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean') &&
            // frontendPrivateIp?: string
            (typeof arg.frontendPrivateIp === 'undefined' || typeof arg.frontendPrivateIp === 'string') &&
            // identityManagement?: IdentityManagementConfig
            (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
            // isPrivateCluster?: boolean
            (typeof arg.isPrivateCluster === 'undefined' || typeof arg.isPrivateCluster === 'boolean') &&
            // kubernetesVersion?: string
            (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
            // labels?: { [key: string]: string }
            (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
            // location?: string
            (typeof arg.location === 'undefined' || typeof arg.location === 'string') &&
            // machineHealthCheckEnabled?: boolean
            (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
            // networking?: TKGNetwork
            (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
            // numOfWorkerNodes?: string
            (typeof arg.numOfWorkerNodes === 'undefined' || typeof arg.numOfWorkerNodes === 'string') &&
            // os?: AzureVirtualMachine
            (typeof arg.os === 'undefined' || isAzureVirtualMachine(arg.os)) &&
            // resourceGroup?: string
            (typeof arg.resourceGroup === 'undefined' || typeof arg.resourceGroup === 'string') &&
            // sshPublicKey?: string
            (typeof arg.sshPublicKey === 'undefined' || typeof arg.sshPublicKey === 'string') &&
            // vnetCidr?: string
            (typeof arg.vnetCidr === 'undefined' || typeof arg.vnetCidr === 'string') &&
            // vnetName?: string
            (typeof arg.vnetName === 'undefined' || typeof arg.vnetName === 'string') &&
            // vnetResourceGroup?: string
            (typeof arg.vnetResourceGroup === 'undefined' || typeof arg.vnetResourceGroup === 'string') &&
            // workerMachineType?: string
            (typeof arg.workerMachineType === 'undefined' || typeof arg.workerMachineType === 'string') &&
            // workerNodeSubnet?: string
            (typeof arg.workerNodeSubnet === 'undefined' || typeof arg.workerNodeSubnet === 'string') &&
            // workerNodeSubnetCidr?: string
            (typeof arg.workerNodeSubnetCidr === 'undefined' || typeof arg.workerNodeSubnetCidr === 'string') &&
            true);
    }
    function isAzureResourceGroup(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // id?: string
            (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
            // location: string
            (typeof arg.location === 'string') &&
            // name: string
            (typeof arg.name === 'string') &&
            true);
    }
    function isAzureSubnet(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // cidr?: string
            (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            true);
    }
    function isAzureVirtualMachine(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // osInfo?: OSInfo
            (typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo)) &&
            true);
    }
    function isAzureVirtualNetwork(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // cidrBlock: string
            (typeof arg.cidrBlock === 'string') &&
            // id?: string
            (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
            // location: string
            (typeof arg.location === 'string') &&
            // name: string
            (typeof arg.name === 'string') &&
            // subnets?: AzureSubnet[]
            (typeof arg.subnets === 'undefined' || (Array.isArray(arg.subnets) && arg.subnets.every(function (item) { return isAzureSubnet(item); }))) &&
            true);
    }
    function isConfigFile(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // filecontents?: string
            (typeof arg.filecontents === 'undefined' || typeof arg.filecontents === 'string') &&
            true);
    }
    function isConfigFileInfo(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // path?: string
            (typeof arg.path === 'undefined' || typeof arg.path === 'string') &&
            true);
    }
    function isDockerDaemonStatus(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // status?: boolean
            (typeof arg.status === 'undefined' || typeof arg.status === 'boolean') &&
            true);
    }
    function isDockerRegionalClusterParams(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // annotations?: { [key: string]: string }
            (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
            // ceipOptIn?: boolean
            (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
            // clusterName?: string
            (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
            // controlPlaneFlavor?: string
            (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
            // identityManagement?: IdentityManagementConfig
            (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
            // kubernetesVersion?: string
            (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
            // labels?: { [key: string]: string }
            (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
            // machineHealthCheckEnabled?: boolean
            (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
            // networking?: TKGNetwork
            (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
            // numOfWorkerNodes?: string
            (typeof arg.numOfWorkerNodes === 'undefined' || typeof arg.numOfWorkerNodes === 'string') &&
            true);
    }
    function isError(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // message?: string
            (typeof arg.message === 'undefined' || typeof arg.message === 'string') &&
            true);
    }
    function isFeatureMap(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // [key: string]: string
            (Object.values(arg).every(function (value) { return typeof value === 'string'; })) &&
            true);
    }
    function isFeatures(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // [key: string]: FeatureMap
            (Object.values(arg).every(function (value) { return isFeatureMap(value); })) &&
            true);
    }
    function isHTTPProxyConfiguration(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // enabled?: boolean
            (typeof arg.enabled === 'undefined' || typeof arg.enabled === 'boolean') &&
            // HTTPProxyPassword?: string
            (typeof arg.HTTPProxyPassword === 'undefined' || typeof arg.HTTPProxyPassword === 'string') &&
            // HTTPProxyURL?: string
            (typeof arg.HTTPProxyURL === 'undefined' || typeof arg.HTTPProxyURL === 'string') &&
            // HTTPProxyUsername?: string
            (typeof arg.HTTPProxyUsername === 'undefined' || typeof arg.HTTPProxyUsername === 'string') &&
            // HTTPSProxyPassword?: string
            (typeof arg.HTTPSProxyPassword === 'undefined' || typeof arg.HTTPSProxyPassword === 'string') &&
            // HTTPSProxyURL?: string
            (typeof arg.HTTPSProxyURL === 'undefined' || typeof arg.HTTPSProxyURL === 'string') &&
            // HTTPSProxyUsername?: string
            (typeof arg.HTTPSProxyUsername === 'undefined' || typeof arg.HTTPSProxyUsername === 'string') &&
            // noProxy?: string
            (typeof arg.noProxy === 'undefined' || typeof arg.noProxy === 'string') &&
            true);
    }
    function isIdentityManagementConfig(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // idm_type: 'oidc' | 'ldap' | 'none'
            (['oidc', 'ldap', 'none'].includes(arg.idm_type)) &&
            // ldap_bind_dn?: string
            (typeof arg.ldap_bind_dn === 'undefined' || typeof arg.ldap_bind_dn === 'string') &&
            // ldap_bind_password?: string
            (typeof arg.ldap_bind_password === 'undefined' || typeof arg.ldap_bind_password === 'string') &&
            // ldap_group_search_base_dn?: string
            (typeof arg.ldap_group_search_base_dn === 'undefined' || typeof arg.ldap_group_search_base_dn === 'string') &&
            // ldap_group_search_filter?: string
            (typeof arg.ldap_group_search_filter === 'undefined' || typeof arg.ldap_group_search_filter === 'string') &&
            // ldap_group_search_group_attr?: string
            (typeof arg.ldap_group_search_group_attr === 'undefined' || typeof arg.ldap_group_search_group_attr === 'string') &&
            // ldap_group_search_name_attr?: string
            (typeof arg.ldap_group_search_name_attr === 'undefined' || typeof arg.ldap_group_search_name_attr === 'string') &&
            // ldap_group_search_user_attr?: string
            (typeof arg.ldap_group_search_user_attr === 'undefined' || typeof arg.ldap_group_search_user_attr === 'string') &&
            // ldap_root_ca?: string
            (typeof arg.ldap_root_ca === 'undefined' || typeof arg.ldap_root_ca === 'string') &&
            // ldap_url?: string
            (typeof arg.ldap_url === 'undefined' || typeof arg.ldap_url === 'string') &&
            // ldap_user_search_base_dn?: string
            (typeof arg.ldap_user_search_base_dn === 'undefined' || typeof arg.ldap_user_search_base_dn === 'string') &&
            // ldap_user_search_email_attr?: string
            (typeof arg.ldap_user_search_email_attr === 'undefined' || typeof arg.ldap_user_search_email_attr === 'string') &&
            // ldap_user_search_filter?: string
            (typeof arg.ldap_user_search_filter === 'undefined' || typeof arg.ldap_user_search_filter === 'string') &&
            // ldap_user_search_id_attr?: string
            (typeof arg.ldap_user_search_id_attr === 'undefined' || typeof arg.ldap_user_search_id_attr === 'string') &&
            // ldap_user_search_name_attr?: string
            (typeof arg.ldap_user_search_name_attr === 'undefined' || typeof arg.ldap_user_search_name_attr === 'string') &&
            // ldap_user_search_username?: string
            (typeof arg.ldap_user_search_username === 'undefined' || typeof arg.ldap_user_search_username === 'string') &&
            // oidc_claim_mappings?: { [key: string]: string }
            (typeof arg.oidc_claim_mappings === 'undefined' || typeof arg.oidc_claim_mappings === 'string') &&
            // oidc_client_id?: string
            (typeof arg.oidc_client_id === 'undefined' || typeof arg.oidc_client_id === 'string') &&
            // oidc_client_secret?: string
            (typeof arg.oidc_client_secret === 'undefined' || typeof arg.oidc_client_secret === 'string') &&
            // oidc_provider_name?: string
            (typeof arg.oidc_provider_name === 'undefined' || typeof arg.oidc_provider_name === 'string') &&
            // oidc_provider_url?: string
            (typeof arg.oidc_provider_url === 'undefined' || typeof arg.oidc_provider_url === 'string') &&
            // oidc_scope?: string
            (typeof arg.oidc_scope === 'undefined' || typeof arg.oidc_scope === 'string') &&
            // oidc_skip_verify_cert?: boolean
            (typeof arg.oidc_skip_verify_cert === 'undefined' || typeof arg.oidc_skip_verify_cert === 'boolean') &&
            true);
    }
    function isLdapParams(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // ldap_bind_dn?: string
            (typeof arg.ldap_bind_dn === 'undefined' || typeof arg.ldap_bind_dn === 'string') &&
            // ldap_bind_password?: string
            (typeof arg.ldap_bind_password === 'undefined' || typeof arg.ldap_bind_password === 'string') &&
            // ldap_group_search_base_dn?: string
            (typeof arg.ldap_group_search_base_dn === 'undefined' || typeof arg.ldap_group_search_base_dn === 'string') &&
            // ldap_group_search_filter?: string
            (typeof arg.ldap_group_search_filter === 'undefined' || typeof arg.ldap_group_search_filter === 'string') &&
            // ldap_group_search_group_attr?: string
            (typeof arg.ldap_group_search_group_attr === 'undefined' || typeof arg.ldap_group_search_group_attr === 'string') &&
            // ldap_group_search_name_attr?: string
            (typeof arg.ldap_group_search_name_attr === 'undefined' || typeof arg.ldap_group_search_name_attr === 'string') &&
            // ldap_group_search_user_attr?: string
            (typeof arg.ldap_group_search_user_attr === 'undefined' || typeof arg.ldap_group_search_user_attr === 'string') &&
            // ldap_root_ca?: string
            (typeof arg.ldap_root_ca === 'undefined' || typeof arg.ldap_root_ca === 'string') &&
            // ldap_test_group?: string
            (typeof arg.ldap_test_group === 'undefined' || typeof arg.ldap_test_group === 'string') &&
            // ldap_test_user?: string
            (typeof arg.ldap_test_user === 'undefined' || typeof arg.ldap_test_user === 'string') &&
            // ldap_url?: string
            (typeof arg.ldap_url === 'undefined' || typeof arg.ldap_url === 'string') &&
            // ldap_user_search_base_dn?: string
            (typeof arg.ldap_user_search_base_dn === 'undefined' || typeof arg.ldap_user_search_base_dn === 'string') &&
            // ldap_user_search_email_attr?: string
            (typeof arg.ldap_user_search_email_attr === 'undefined' || typeof arg.ldap_user_search_email_attr === 'string') &&
            // ldap_user_search_filter?: string
            (typeof arg.ldap_user_search_filter === 'undefined' || typeof arg.ldap_user_search_filter === 'string') &&
            // ldap_user_search_id_attr?: string
            (typeof arg.ldap_user_search_id_attr === 'undefined' || typeof arg.ldap_user_search_id_attr === 'string') &&
            // ldap_user_search_name_attr?: string
            (typeof arg.ldap_user_search_name_attr === 'undefined' || typeof arg.ldap_user_search_name_attr === 'string') &&
            // ldap_user_search_username?: string
            (typeof arg.ldap_user_search_username === 'undefined' || typeof arg.ldap_user_search_username === 'string') &&
            true);
    }
    function isLdapTestResult(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // code?: number
            (typeof arg.code === 'undefined' || typeof arg.code === 'number') &&
            // desc?: string
            (typeof arg.desc === 'undefined' || typeof arg.desc === 'string') &&
            true);
    }
    function isNodeType(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // cpu?: number
            (typeof arg.cpu === 'undefined' || typeof arg.cpu === 'number') &&
            // disk?: number
            (typeof arg.disk === 'undefined' || typeof arg.disk === 'number') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // ram?: number
            (typeof arg.ram === 'undefined' || typeof arg.ram === 'number') &&
            true);
    }
    function isOSInfo(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // arch?: string
            (typeof arg.arch === 'undefined' || typeof arg.arch === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // version?: string
            (typeof arg.version === 'undefined' || typeof arg.version === 'string') &&
            true);
    }
    function isProviderInfo(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // provider?: string
            (typeof arg.provider === 'undefined' || typeof arg.provider === 'string') &&
            // tkrVersion?: string
            (typeof arg.tkrVersion === 'undefined' || typeof arg.tkrVersion === 'string') &&
            true);
    }
    function isTKGNetwork(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // clusterDNSName?: string
            (typeof arg.clusterDNSName === 'undefined' || typeof arg.clusterDNSName === 'string') &&
            // clusterNodeCIDR?: string
            (typeof arg.clusterNodeCIDR === 'undefined' || typeof arg.clusterNodeCIDR === 'string') &&
            // clusterPodCIDR?: string
            (typeof arg.clusterPodCIDR === 'undefined' || typeof arg.clusterPodCIDR === 'string') &&
            // clusterServiceCIDR?: string
            (typeof arg.clusterServiceCIDR === 'undefined' || typeof arg.clusterServiceCIDR === 'string') &&
            // cniType?: string
            (typeof arg.cniType === 'undefined' || typeof arg.cniType === 'string') &&
            // httpProxyConfiguration?: HTTPProxyConfiguration
            (typeof arg.httpProxyConfiguration === 'undefined' || isHTTPProxyConfiguration(arg.httpProxyConfiguration)) &&
            // networkName?: string
            (typeof arg.networkName === 'undefined' || typeof arg.networkName === 'string') &&
            true);
    }
    function isVpc(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // cidr?: string
            (typeof arg.cidr === 'undefined' || typeof arg.cidr === 'string') &&
            // id?: string
            (typeof arg.id === 'undefined' || typeof arg.id === 'string') &&
            true);
    }
    function isVSphereAvailabilityZone(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // moid?: string
            (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            true);
    }
    function isVSphereCredentials(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // host?: string
            (typeof arg.host === 'undefined' || typeof arg.host === 'string') &&
            // insecure?: boolean
            (typeof arg.insecure === 'undefined' || typeof arg.insecure === 'boolean') &&
            // password?: string
            (typeof arg.password === 'undefined' || typeof arg.password === 'string') &&
            // thumbprint?: string
            (typeof arg.thumbprint === 'undefined' || typeof arg.thumbprint === 'string') &&
            // username?: string
            (typeof arg.username === 'undefined' || typeof arg.username === 'string') &&
            true);
    }
    function isVSphereDatacenter(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // moid?: string
            (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            true);
    }
    function isVSphereDatastore(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // moid?: string
            (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            true);
    }
    function isVSphereFolder(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // moid?: string
            (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            true);
    }
    function isVsphereInfo(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // hasPacific?: string
            (typeof arg.hasPacific === 'undefined' || typeof arg.hasPacific === 'string') &&
            // version?: string
            (typeof arg.version === 'undefined' || typeof arg.version === 'string') &&
            true);
    }
    function isVSphereManagementObject(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // moid?: string
            (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // parentMoid?: string
            (typeof arg.parentMoid === 'undefined' || typeof arg.parentMoid === 'string') &&
            // path?: string
            (typeof arg.path === 'undefined' || typeof arg.path === 'string') &&
            // resourceType?: 'datacenter' | 'cluster' | 'hostgroup' | 'folder' | 'respool' | 'vm' | 'datastore' | 'host' | 'network'
            (typeof arg.resourceType === 'undefined' || ['datacenter', 'cluster', 'hostgroup', 'folder', 'respool', 'vm', 'datastore', 'host', 'network'].includes(arg.resourceType)) &&
            true);
    }
    function isVSphereNetwork(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // displayName?: string
            (typeof arg.displayName === 'undefined' || typeof arg.displayName === 'string') &&
            // moid?: string
            (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            true);
    }
    function isVSphereRegion(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // moid?: string
            (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // zones?: VSphereAvailabilityZone[]
            (typeof arg.zones === 'undefined' || (Array.isArray(arg.zones) && arg.zones.every(function (item) { return isVSphereAvailabilityZone(item); }))) &&
            true);
    }
    function isVsphereRegionalClusterParams(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // annotations?: { [key: string]: string }
            (typeof arg.annotations === 'undefined' || typeof arg.annotations === 'string') &&
            // aviConfig?: AviConfig
            (typeof arg.aviConfig === 'undefined' || isAviConfig(arg.aviConfig)) &&
            // ceipOptIn?: boolean
            (typeof arg.ceipOptIn === 'undefined' || typeof arg.ceipOptIn === 'boolean') &&
            // clusterName?: string
            (typeof arg.clusterName === 'undefined' || typeof arg.clusterName === 'string') &&
            // controlPlaneEndpoint?: string
            (typeof arg.controlPlaneEndpoint === 'undefined' || typeof arg.controlPlaneEndpoint === 'string') &&
            // controlPlaneFlavor?: string
            (typeof arg.controlPlaneFlavor === 'undefined' || typeof arg.controlPlaneFlavor === 'string') &&
            // controlPlaneNodeType?: string
            (typeof arg.controlPlaneNodeType === 'undefined' || typeof arg.controlPlaneNodeType === 'string') &&
            // datacenter?: string
            (typeof arg.datacenter === 'undefined' || typeof arg.datacenter === 'string') &&
            // datastore?: string
            (typeof arg.datastore === 'undefined' || typeof arg.datastore === 'string') &&
            // enableAuditLogging?: boolean
            (typeof arg.enableAuditLogging === 'undefined' || typeof arg.enableAuditLogging === 'boolean') &&
            // folder?: string
            (typeof arg.folder === 'undefined' || typeof arg.folder === 'string') &&
            // identityManagement?: IdentityManagementConfig
            (typeof arg.identityManagement === 'undefined' || isIdentityManagementConfig(arg.identityManagement)) &&
            // ipFamily?: string
            (typeof arg.ipFamily === 'undefined' || typeof arg.ipFamily === 'string') &&
            // kubernetesVersion?: string
            (typeof arg.kubernetesVersion === 'undefined' || typeof arg.kubernetesVersion === 'string') &&
            // labels?: { [key: string]: string }
            (typeof arg.labels === 'undefined' || typeof arg.labels === 'string') &&
            // machineHealthCheckEnabled?: boolean
            (typeof arg.machineHealthCheckEnabled === 'undefined' || typeof arg.machineHealthCheckEnabled === 'boolean') &&
            // networking?: TKGNetwork
            (typeof arg.networking === 'undefined' || isTKGNetwork(arg.networking)) &&
            // numOfWorkerNode?: number
            (typeof arg.numOfWorkerNode === 'undefined' || typeof arg.numOfWorkerNode === 'number') &&
            // os?: VSphereVirtualMachine
            (typeof arg.os === 'undefined' || isVSphereVirtualMachine(arg.os)) &&
            // resourcePool?: string
            (typeof arg.resourcePool === 'undefined' || typeof arg.resourcePool === 'string') &&
            // ssh_key?: string
            (typeof arg.ssh_key === 'undefined' || typeof arg.ssh_key === 'string') &&
            // vsphereCredentials?: VSphereCredentials
            (typeof arg.vsphereCredentials === 'undefined' || isVSphereCredentials(arg.vsphereCredentials)) &&
            // workerNodeType?: string
            (typeof arg.workerNodeType === 'undefined' || typeof arg.workerNodeType === 'string') &&
            true);
    }
    function isVSphereResourcePool(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // moid?: string
            (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            true);
    }
    function isVSphereThumbprint(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // insecure?: boolean
            (typeof arg.insecure === 'undefined' || typeof arg.insecure === 'boolean') &&
            // thumbprint?: string
            (typeof arg.thumbprint === 'undefined' || typeof arg.thumbprint === 'string') &&
            true);
    }
    function isVSphereVirtualMachine(arg) {
        return (arg != null &&
            typeof arg === 'object' &&
            // isTemplate: boolean
            (typeof arg.isTemplate === 'boolean') &&
            // k8sVersion?: string
            (typeof arg.k8sVersion === 'undefined' || typeof arg.k8sVersion === 'string') &&
            // moid?: string
            (typeof arg.moid === 'undefined' || typeof arg.moid === 'string') &&
            // name?: string
            (typeof arg.name === 'undefined' || typeof arg.name === 'string') &&
            // osInfo?: OSInfo
            (typeof arg.osInfo === 'undefined' || isOSInfo(arg.osInfo)) &&
            true);
    }

    /**
     * Created with https://github.com/flowup/api-client-generator
     */
    var GuardedAPIClient = /** @class */ (function (_super) {
        __extends(GuardedAPIClient, _super);
        function GuardedAPIClient(httpClient, domain, options) {
            var _this = _super.call(this, httpClient, domain, options) || this;
            _this.httpClient = httpClient;
            return _this;
        }
        GuardedAPIClient.prototype.getUI = function (requestHttpOptions) {
            return _super.prototype.getUI.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isFile(res) || console.error("TypeGuard for response 'File' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getUIFile = function (args, requestHttpOptions) {
            return _super.prototype.getUIFile.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isFile(res) || console.error("TypeGuard for response 'File' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getFeatureFlags = function (requestHttpOptions) {
            return _super.prototype.getFeatureFlags.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isFeatures(res) || console.error("TypeGuard for response 'Features' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getTanzuEdition = function (requestHttpOptions) {
            return _super.prototype.getTanzuEdition.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.verifyLdapConnect = function (args, requestHttpOptions) {
            return _super.prototype.verifyLdapConnect.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isLdapTestResult(res) || console.error("TypeGuard for response 'LdapTestResult' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.verifyLdapBind = function (requestHttpOptions) {
            return _super.prototype.verifyLdapBind.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isLdapTestResult(res) || console.error("TypeGuard for response 'LdapTestResult' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.verifyLdapUserSearch = function (requestHttpOptions) {
            return _super.prototype.verifyLdapUserSearch.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isLdapTestResult(res) || console.error("TypeGuard for response 'LdapTestResult' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.verifyLdapGroupSearch = function (requestHttpOptions) {
            return _super.prototype.verifyLdapGroupSearch.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isLdapTestResult(res) || console.error("TypeGuard for response 'LdapTestResult' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAviClouds = function (requestHttpOptions) {
            return _super.prototype.getAviClouds.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAviCloud(res) || console.error("TypeGuard for response 'AviCloud' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAviServiceEngineGroups = function (requestHttpOptions) {
            return _super.prototype.getAviServiceEngineGroups.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAviServiceEngineGroup(res) || console.error("TypeGuard for response 'AviServiceEngineGroup' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAviVipNetworks = function (requestHttpOptions) {
            return _super.prototype.getAviVipNetworks.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAviVipNetwork(res) || console.error("TypeGuard for response 'AviVipNetwork' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getProvider = function (requestHttpOptions) {
            return _super.prototype.getProvider.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isProviderInfo(res) || console.error("TypeGuard for response 'ProviderInfo' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getVsphereThumbprint = function (args, requestHttpOptions) {
            return _super.prototype.getVsphereThumbprint.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isVSphereThumbprint(res) || console.error("TypeGuard for response 'VSphereThumbprint' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.setVSphereEndpoint = function (args, requestHttpOptions) {
            return _super.prototype.setVSphereEndpoint.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isVsphereInfo(res) || console.error("TypeGuard for response 'VsphereInfo' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getVSphereDatacenters = function (requestHttpOptions) {
            return _super.prototype.getVSphereDatacenters.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isVSphereDatacenter(res) || console.error("TypeGuard for response 'VSphereDatacenter' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getVSphereDatastores = function (args, requestHttpOptions) {
            return _super.prototype.getVSphereDatastores.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isVSphereDatastore(res) || console.error("TypeGuard for response 'VSphereDatastore' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getVSphereFolders = function (args, requestHttpOptions) {
            return _super.prototype.getVSphereFolders.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isVSphereFolder(res) || console.error("TypeGuard for response 'VSphereFolder' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getVSphereComputeResources = function (args, requestHttpOptions) {
            return _super.prototype.getVSphereComputeResources.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isVSphereManagementObject(res) || console.error("TypeGuard for response 'VSphereManagementObject' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getVSphereResourcePools = function (args, requestHttpOptions) {
            return _super.prototype.getVSphereResourcePools.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isVSphereResourcePool(res) || console.error("TypeGuard for response 'VSphereResourcePool' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getVSphereNetworks = function (args, requestHttpOptions) {
            return _super.prototype.getVSphereNetworks.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isVSphereNetwork(res) || console.error("TypeGuard for response 'VSphereNetwork' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getVSphereNodeTypes = function (requestHttpOptions) {
            return _super.prototype.getVSphereNodeTypes.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isNodeType(res) || console.error("TypeGuard for response 'NodeType' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getVSphereOSImages = function (args, requestHttpOptions) {
            return _super.prototype.getVSphereOSImages.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isVSphereVirtualMachine(res) || console.error("TypeGuard for response 'VSphereVirtualMachine' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.exportTKGConfigForVsphere = function (args, requestHttpOptions) {
            return _super.prototype.exportTKGConfigForVsphere.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.applyTKGConfigForVsphere = function (args, requestHttpOptions) {
            return _super.prototype.applyTKGConfigForVsphere.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isConfigFileInfo(res) || console.error("TypeGuard for response 'ConfigFileInfo' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.importTKGConfigForVsphere = function (args, requestHttpOptions) {
            return _super.prototype.importTKGConfigForVsphere.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isVsphereRegionalClusterParams(res) || console.error("TypeGuard for response 'VsphereRegionalClusterParams' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.createVSphereRegionalCluster = function (args, requestHttpOptions) {
            return _super.prototype.createVSphereRegionalCluster.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getVPCs = function (requestHttpOptions) {
            return _super.prototype.getVPCs.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isVpc(res) || console.error("TypeGuard for response 'Vpc' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAWSNodeTypes = function (args, requestHttpOptions) {
            return _super.prototype.getAWSNodeTypes.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAWSRegions = function (requestHttpOptions) {
            return _super.prototype.getAWSRegions.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAWSOSImages = function (args, requestHttpOptions) {
            return _super.prototype.getAWSOSImages.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAWSVirtualMachine(res) || console.error("TypeGuard for response 'AWSVirtualMachine' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAWSCredentialProfiles = function (requestHttpOptions) {
            return _super.prototype.getAWSCredentialProfiles.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAWSAvailabilityZones = function (requestHttpOptions) {
            return _super.prototype.getAWSAvailabilityZones.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAWSAvailabilityZone(res) || console.error("TypeGuard for response 'AWSAvailabilityZone' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAWSSubnets = function (args, requestHttpOptions) {
            return _super.prototype.getAWSSubnets.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAWSSubnet(res) || console.error("TypeGuard for response 'AWSSubnet' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.exportTKGConfigForAWS = function (args, requestHttpOptions) {
            return _super.prototype.exportTKGConfigForAWS.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.applyTKGConfigForAWS = function (args, requestHttpOptions) {
            return _super.prototype.applyTKGConfigForAWS.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isConfigFileInfo(res) || console.error("TypeGuard for response 'ConfigFileInfo' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.createAWSRegionalCluster = function (args, requestHttpOptions) {
            return _super.prototype.createAWSRegionalCluster.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.importTKGConfigForAWS = function (args, requestHttpOptions) {
            return _super.prototype.importTKGConfigForAWS.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAWSRegionalClusterParams(res) || console.error("TypeGuard for response 'AWSRegionalClusterParams' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAzureEndpoint = function (requestHttpOptions) {
            return _super.prototype.getAzureEndpoint.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAzureAccountParams(res) || console.error("TypeGuard for response 'AzureAccountParams' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAzureResourceGroups = function (args, requestHttpOptions) {
            return _super.prototype.getAzureResourceGroups.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAzureResourceGroup(res) || console.error("TypeGuard for response 'AzureResourceGroup' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.createAzureResourceGroup = function (args, requestHttpOptions) {
            return _super.prototype.createAzureResourceGroup.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAzureVnets = function (args, requestHttpOptions) {
            return _super.prototype.getAzureVnets.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAzureVirtualNetwork(res) || console.error("TypeGuard for response 'AzureVirtualNetwork' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.createAzureVirtualNetwork = function (args, requestHttpOptions) {
            return _super.prototype.createAzureVirtualNetwork.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAzureOSImages = function (requestHttpOptions) {
            return _super.prototype.getAzureOSImages.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAzureVirtualMachine(res) || console.error("TypeGuard for response 'AzureVirtualMachine' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAzureRegions = function (requestHttpOptions) {
            return _super.prototype.getAzureRegions.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAzureLocation(res) || console.error("TypeGuard for response 'AzureLocation' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.getAzureInstanceTypes = function (args, requestHttpOptions) {
            return _super.prototype.getAzureInstanceTypes.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAzureInstanceType(res) || console.error("TypeGuard for response 'AzureInstanceType' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.exportTKGConfigForAzure = function (args, requestHttpOptions) {
            return _super.prototype.exportTKGConfigForAzure.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.applyTKGConfigForAzure = function (args, requestHttpOptions) {
            return _super.prototype.applyTKGConfigForAzure.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isConfigFileInfo(res) || console.error("TypeGuard for response 'ConfigFileInfo' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.createAzureRegionalCluster = function (args, requestHttpOptions) {
            return _super.prototype.createAzureRegionalCluster.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.importTKGConfigForAzure = function (args, requestHttpOptions) {
            return _super.prototype.importTKGConfigForAzure.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isAzureRegionalClusterParams(res) || console.error("TypeGuard for response 'AzureRegionalClusterParams' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.checkIfDockerDaemonAvailable = function (requestHttpOptions) {
            return _super.prototype.checkIfDockerDaemonAvailable.call(this, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isDockerDaemonStatus(res) || console.error("TypeGuard for response 'DockerDaemonStatus' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.exportTKGConfigForDocker = function (args, requestHttpOptions) {
            return _super.prototype.exportTKGConfigForDocker.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.applyTKGConfigForDocker = function (args, requestHttpOptions) {
            return _super.prototype.applyTKGConfigForDocker.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isConfigFileInfo(res) || console.error("TypeGuard for response 'ConfigFileInfo' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.createDockerRegionalCluster = function (args, requestHttpOptions) {
            return _super.prototype.createDockerRegionalCluster.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return typeof res === 'string' || console.error("TypeGuard for response 'string' caught inconsistency.", res); }));
        };
        GuardedAPIClient.prototype.importTKGConfigForDocker = function (args, requestHttpOptions) {
            return _super.prototype.importTKGConfigForDocker.call(this, args, requestHttpOptions)
                .pipe(operators.tap(function (res) { return isDockerRegionalClusterParams(res) || console.error("TypeGuard for response 'DockerRegionalClusterParams' caught inconsistency.", res); }));
        };
        return GuardedAPIClient;
    }(APIClient));
    GuardedAPIClient.ɵfac = i0__namespace.ɵɵngDeclareFactory({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0__namespace, type: GuardedAPIClient, deps: [{ token: i1__namespace.HttpClient }, { token: USE_DOMAIN, optional: true }, { token: USE_HTTP_OPTIONS, optional: true }], target: i0__namespace.ɵɵFactoryTarget.Injectable });
    GuardedAPIClient.ɵprov = i0__namespace.ɵɵngDeclareInjectable({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0__namespace, type: GuardedAPIClient });
    i0__namespace.ɵɵngDeclareClassMetadata({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0__namespace, type: GuardedAPIClient, decorators: [{
                type: i0.Injectable
            }], ctorParameters: function () {
            return [{ type: i1__namespace.HttpClient }, { type: undefined, decorators: [{
                            type: i0.Optional
                        }, {
                            type: i0.Inject,
                            args: [USE_DOMAIN]
                        }] }, { type: undefined, decorators: [{
                            type: i0.Optional
                        }, {
                            type: i0.Inject,
                            args: [USE_HTTP_OPTIONS]
                        }] }];
        } });

    var APIClientModule = /** @class */ (function () {
        function APIClientModule() {
        }
        /**
         * Use this method in your root module to provide the APIClientModule
         *
         * If you are not providing
         * @param { APIClientModuleConfig } config
         * @returns { ModuleWithProviders }
         */
        APIClientModule.forRoot = function (config) {
            if (config === void 0) { config = {}; }
            return {
                ngModule: APIClientModule,
                providers: __spreadArray(__spreadArray(__spreadArray([], __read((config.domain != null ? [{ provide: USE_DOMAIN, useValue: config.domain }] : []))), __read((config.httpOptions ? [{ provide: USE_HTTP_OPTIONS, useValue: config.httpOptions }] : []))), __read((config.guardResponses ? [{ provide: APIClient, useClass: GuardedAPIClient }] : [APIClient])))
            };
        };
        return APIClientModule;
    }());
    APIClientModule.ɵfac = i0__namespace.ɵɵngDeclareFactory({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0__namespace, type: APIClientModule, deps: [], target: i0__namespace.ɵɵFactoryTarget.NgModule });
    APIClientModule.ɵmod = i0__namespace.ɵɵngDeclareNgModule({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0__namespace, type: APIClientModule });
    APIClientModule.ɵinj = i0__namespace.ɵɵngDeclareInjector({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0__namespace, type: APIClientModule });
    i0__namespace.ɵɵngDeclareClassMetadata({ minVersion: "12.0.0", version: "12.2.15", ngImport: i0__namespace, type: APIClientModule, decorators: [{
                type: i0.NgModule,
                args: [{}]
            }] });

    /*
     * Public API Surface of tanzu-management-cluster-ng-api
     * Exports swagger generated APIClient and modules, and swagger generated models
     */

    /**
     * Generated bundle index. Do not edit.
     */

    exports.APIClient = APIClient;
    exports.APIClientModule = APIClientModule;
    exports.GuardedAPIClient = GuardedAPIClient;

    Object.defineProperty(exports, '__esModule', { value: true });

}));
//# sourceMappingURL=tanzu-management-cluster-ng-api.umd.js.map
