(function() {
  if (!window.console) {
    window.console = {
      log: (function(obj) {})
    };
  }

  $(function() {
    var getIOSVersion, introAnimationRef, jsonPostLoad, mobileFixNavItem;
    window.requestAnimationFrame = window.requestAnimationFrame || window.mozRequestAnimationFrame || window.webkitRequestAnimationFrame || window.msRequestAnimationFrame;
    getIOSVersion = function() {
      var v;
      if (/iP(hone|od|ad)/.test(navigator.platform)) {
        v = navigator.appVersion.match(/OS (\d+)_(\d+)_?(\d+)?/);
        return [parseInt(v[1], 10), parseInt(v[2], 10), parseInt(v[3] || 0, 10)];
      } else {
        return false;
      }
    };
    window.has_ios = getIOSVersion();
    window.is_mobile = navigator.userAgent.match(/iPhone/i) || navigator.userAgent.match(/iPod/i) || navigator.userAgent.match(/iPad/i) || navigator.userAgent.match(/Android/i) ? true : false;
    window.is_iphone = navigator.userAgent.match(/iPhone/i) || navigator.userAgent.match(/iPod/i) ? true : false;
    window.is_ipad = navigator.userAgent.match(/iPad/i) ? true : false;
    if (window.is_mobile) {
      $('body').addClass('deviceMobile');
    }
    window.breakpointer = {
      s: 290,
      m: 650,
      l: 959
    };
    window.transEndEventNames = {
      WebkitTransition: "webkitTransitionEnd",
      MozTransition: "transitionend",
      transition: "transitionend"
    };
    window.transEndEventName = window.transEndEventNames[Modernizr.prefixed("transition")];
    Modernizr.addTest("highres", function() {
      var dpr;
      dpr = window.devicePixelRatio || (window.screen.deviceXDPI / window.screen.logicalXDPI) || 1;
      return !!(dpr > 1);
    });
    if (typeof Modernizr.prefixed("transform") === 'string') {
      window.prefixedTransform = Modernizr.prefixed("transform").replace(/([A-Z])/g, function(str, m1) {
        return "-" + m1.toLowerCase();
      }).replace(/^ms-/, "-ms-");
    }
    window.helpers = {
      detectIE8: (function(_this) {
        return function() {
          var IE_version, msie, ua;
          ua = window.navigator.userAgent;
          msie = ua.indexOf('MSIE ');
          IE_version = 0;
          if (msie > 0) {
            IE_version = parseInt(ua.substring(msie + 5, ua.indexOf('.', msie)), 10);
          }
          if (IE_version !== 0 && IE_version < 9) {
            return true;
          } else {
            return false;
          }
        };
      })(this)
    };
    introAnimationRef = $('.home');
    if (introAnimationRef.length > 0) {
      new introEffect(introAnimationRef);
    }
    jsonPostLoad = $('.js-post-load');
    if (jsonPostLoad.length > 0) {
      new loadPost(jsonPostLoad);
    }
    if (window.is_mobile) {
      mobileFixNavItem = $('.js-mobileFixNavItem');
      if (mobileFixNavItem.length > 0) {
        return new mobileFixNav(mobileFixNavItem);
      }
    }
  });

}).call(this);

(function() {
  var __bind = function(fn, me){ return function(){ return fn.apply(me, arguments); }; };

  window.heightProgressBar = (function() {
    function heightProgressBar(ref) {
      this.ref = ref;
      this.onScroll = __bind(this.onScroll, this);
      this.body = $('body');
      this.auxBody = this.body.find('.auxBody');
      this.auxBody.on('scroll', this.onScroll);
      this.auxBody_h = this.auxBody.height();
      this.content = this.body.find('.main');
      this.content_h = this.content.height();
      this.scrollOffset = this.content_h - this.auxBody_h;
    }

    heightProgressBar.prototype.onScroll = function() {
      return this.scrollTop = this.auxBody.scrollTop();
    };

    return heightProgressBar;

  })();

}).call(this);

(function() {
  var __bind = function(fn, me){ return function(){ return fn.apply(me, arguments); }; };

  window.introEffect = (function() {
    function introEffect(ref) {
      this.ref = ref;
      this.enable_scroll = __bind(this.enable_scroll, this);
      this.disable_scroll = __bind(this.disable_scroll, this);
      this.onScroll = __bind(this.onScroll, this);
      this.window_ref = $(window);
      this.body = $('body');
      this.auxBody = this.body.find('.auxBody');
      this.auxBody.on('scroll', this.onScroll);
      this.container = this.body.find('.wrapper');
      this.animate = false;
    }

    introEffect.prototype.onScroll = function() {
      console.log('trace');
      this.scrollTop = this.auxBody.scrollTop();
      if (this.animate === false && this.scrollTop >= 1) {
        this.body.addClass('fixed');
        this.disable_scroll();
      } else if (this.scrollTop === 0) {
        this.body.removeClass('fixed');
        this.animate = false;
      }
    };

    introEffect.prototype.disable_scroll = function() {
      this.auxBody.addClass('stop-scrolling');
      this.animate = true;
      return this.enable_scroll();
    };

    introEffect.prototype.enable_scroll = function() {
      var tmOut;
      return tmOut = setTimeout((function(_this) {
        return function() {
          _this.auxBody.removeClass('stop-scrolling');
          return clearTimeout(tmOut);
        };
      })(this), 1000);
    };

    return introEffect;

  })();

}).call(this);

(function() {
  var __bind = function(fn, me){ return function(){ return fn.apply(me, arguments); }; };

  window.loadPost = (function() {
    function loadPost(ref) {
      this.ref = ref;
      this.onDataLoaded = __bind(this.onDataLoaded, this);
      this.loadJson = __bind(this.loadJson, this);
      this.json_url = this.ref.attr("data-url");
      this.loadJson();
    }

    loadPost.prototype.loadJson = function() {
      return $.getJSON(this.json_url, this.onDataLoaded);
    };

    loadPost.prototype.onDataLoaded = function(data) {
      var htmlContent, i, id, posts, _i, _ref, _results;
      posts = data.post;
      _results = [];
      for (i = _i = 0, _ref = posts.length; 0 <= _ref ? _i < _ref : _i > _ref; i = 0 <= _ref ? ++_i : --_i) {
        id = posts[i].id;
        htmlContent = posts[i].content;
        _results.push(this.ref.append(htmlContent));
      }
      return _results;
    };

    return loadPost;

  })();

}).call(this);

(function() {
  var __bind = function(fn, me){ return function(){ return fn.apply(me, arguments); }; };

  window.mobileFixNav = (function() {
    function mobileFixNav(ref) {
      this.ref = ref;
      this.onScroll = __bind(this.onScroll, this);
      this.body = $('body');
      this.navBar = this.ref.find('.navbar');
      this.auxBody = this.body.find('.auxBody');
      this.auxBody.on('scroll', this.onScroll);
    }

    mobileFixNav.prototype.onScroll = function() {
      this.refOffset = this.ref.offset().top;
      this.navOffset = this.navBar.offset().top;
      console.log(this.refOffset, this.navOffset);
      if (this.refOffset <= 0) {
        return this.ref.addClass('menuFixed');
      } else {
        return this.ref.removeClass('menuFixed');
      }
    };

    return mobileFixNav;

  })();

}).call(this);
