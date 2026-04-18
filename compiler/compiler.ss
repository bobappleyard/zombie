#!/usr/bin/csi -ss

(import srfi-1)
(import srfi-111)
(import srfi-152)

(define (main args)
  (let ([pkg (make-package '() '() 0 (list (make-block 0 0 '())))])
    (let next ()
      (define tl (read))
      (if (eof-object? tl)
        (print pkg)
        (begin
          (compile-toplevel pkg tl)
          (next))))))

(define (compile-toplevel pkg tl)
  (define (compile-expr expr)
    (expr->instructions
      (expr->anf
        (closures->currying
          (mutation->boxing `((lamdda () ,expr)) '())
          '()))
      (make-scope pkg '() #f (box 0))))
  (cond
    [(null? tl) (error "empty list is not a valid expression")]
    [(atom? tl)]
    [else       (case (car tl)
                  [(define) (let ([b (car (package-blocks pkg))])
                              (block-code-set! b
                                               (append (block-code b)
                                                       (compile-expr (caddr tl))
                                                       `(global-define ,(package-global! pkg (cadr tl)))
                                                       `(global-set ,(package-global! pkg (cadr tl))))))]
                  [else (error "unsupported" tl)])]))

(define (mutation->boxing e vs)
  (define (enter-scope vars body)
    (let* ([boxed (find-vars boxed-var? body)]
           [need (filter (lambda (v) (member v boxed)) vars)]
           [inner (append (filter (lambda (v) (not (member v vars))) vs)
                          need)])
      (append (map (lambda (v) `(set! ,v (make-box ,v))) need)
              (map (lambda (e) (mutation->boxing e inner)) body))))
  (cond
    [(and (symbol? e) 
          (member e vs)) `(box-ref ,e)]
    [(atom? e)           e]
    [else                (case (car e)
                           [(begin if) (cons (car e) (map (lambda (e) (mutation->boxing e vs)) (cdr e)))]
                           [(set!)     `(,(if (member (cadr e) vs) 'box-set! 'set!) ,(cadr e) ,(mutation->boxing (caddr e) vs))]
                           [(lambda)   `(lambda ,(cadr e) ,@(enter-scope (cadr e) (cddr e)))]
                           [(let)      `(let ,(map (lambda (b) `(,(car b) ,(mutation->boxing (cadr b) vs)))
                                                   (cadr e))
                                         ,@(enter-scope (map car (cadr e)) (cddr e)))]
                           [else       (map (lambda (e) (mutation->boxing e vs)) e)])]))

(define (closures->currying e vs)
  (cond
    [(atom? e) e]
    [else (case (car e)
       [(begin if set!) (cons (car e) (map (lambda (e) (closures->currying e vs)) (cdr e)))]
       [(lambda)        (let* ([free (find-vars free-var? e)]
                               [need (filter (lambda (v) (member v free))
                                             vs)]
                               [inner (append need (cadr e))]
                               [body (map (lambda (e) (closures->currying e inner)) (cddr e))])
                          (if (null? need)
                            `(lambda ,inner ,@body)
                            `(curry (lambda ,inner ,@body) ,@need)))]
       [(let)           (let ([inner (append (filter (lambda (v) (not (member v vs))) (map car (cadr e)))
                                             vs)])
                          `(let ,(map (lambda (b) `(,(car b) ,(closures->currying (cadr b) vs))) (cadr e))
                            ,@(map (lambda (e) (closures->currying e inner)) (cddr e))))]
       [else            (map (lambda (e) (closures->currying e vs)) e)])]))

(define (find-vars ep expr)
  (delete-duplicates!
    (let next ([expr expr] [scope '()] [ctx '()])
      (cond
        [(and (symbol? expr)
              (ep expr scope ctx)) (list expr)]
        [(atom? expr)              '()]
        [else                      (let ([ctx (cons expr ctx)])
                                     (define (inner ext body)
                                       (let ([scope (append ext scope)])
                                         (append-map! (lambda (expr) (next expr scope ctx))
                                                       body)))
                                     (case (car expr) 
                                       [(lambda) (inner (cadr expr) (cddr expr))]
                                       [(let)    (append (inner '() (map cadr (cadr expr)))
                                                         (inner (map car (cadr expr)) (cddr expr)))]
                                       [else     (inner '() expr)]))]))))

(define (free-var? v scope ctx)
  (not (member v scope)))

(define (boxed-var? v scope ctx)
  (and (not (null? ctx))
       (eqv? 'set! (caar ctx))
       (eqv? v (cadar ctx))
       (member 'lambda (map car ctx))
       (free-var? v scope ctx)))

(define (slots-needed e)
  (cond
    [(atom? e) 0]
    [else      (case (car e)
                 [(begin if set!) (apply max 0 (map slots-needed (cdr e)))]
                 [(lambda)        0]
                 [(let)           (+ (apply max (length (cadr e))
                                                (map + (list-tabulate (length (cadr e)) values)
                                                       (map slots-needed (map cadr (cadr e)))))
                                     (apply max 0 (map slots-needed (cddr e))))]
                 [else            (apply max (map slots-needed e))])]))

(define (expr->anf e)
  (if (atom? e)
    e
    (case (car e)
      [(begin if set!) (map expr->anf e)]
      [(lambda)        `(lambda ,(cadr e)
                         ,@(map expr->anf (cddr e)))]
      [(let)           `(let ,(map (lambda (b)
                                     `(,(car b) ,(expr->anf (cadr b))))
                             (cadr e))
                         ,@(map expr->anf (cddr e)))]
      [else            (call->anf e)])))

(define anf-gensym gensym)

(define (call->anf e)
  (let next ([e e] [call '()] [pre '()])
    (define (simple v)
      (next (cdr e) (cons v call) pre))
    (define (complex s) 
      (let ([v (anf-gensym)])
        (next (cdr e) (cons v call) (cons `(let ([,v ,s])) pre))))
    (cond
      [(null? e) (fold (lambda (x acc) (append x (list acc)))
                       (reverse call)
                       pre)]
      [(atom?    (car e)) (simple (car e))]
      [else      (case (caar e)
                   [(if set! let) (complex (expr->anf (car e)))]
                   [(begin)       (if (= 2 (length (car e)))
                                    (next (cons (cadar e) (cdr e)) call pre)
                                    (complex (expr->anf (car e))))]
                   [(lambda)      (simple (expr->anf (car e)))]
                   [else          (complex (expr->anf (car e)))])])))

(define (render-package pkg)
  (print "#include<zbc.h>")
  (print "Z_GLOBALS_DECL(" (package-globals pkg) ");")
  (for-each (lambda (b) (render-block b)) (package-blocks pkg))
  (render-block (package-init pkg))
  (print "Z_INIT_DECL(" (block-id (package-init pkg)) ");"))

(define (render-block b)
  (print "Z_BLOCK_DECL(" (block-id b) ") {")
  (print "    Z_PROLOG(" (block-argc b) ", " (block-varc b) ");")
  (for-each (lambda (op)
              (apply print 
                     (append (list "    Z_" (format-instruction (car op)) "(")
                             (intersperse (cdr op) ", ")
                             (list ");"))))
            (block-code b))
  (print "    Z_EPILOG();")
  (print "}"))

(define (format-instruction op)
  (string-map (lambda (c)
                (cond
                  [(eqv? c #\-)          #\_]
                  [(and (char>=? c #\a) 
                        (char<=? c #\z)) (integer->char (+ (char->integer #\A) 
                                                           (- (char->integer c) (char->integer #\a))))]
                  [else                  c]))
              (symbol->string op)))

(define-record package globals values init blocks)
(define-record block argc varc code)
(define-record scope pkg locals tail-position? point-ref)

(define (expr->instructions expr scope)
  (define (value instr)
    (append (list instr) (if (scope-tail-position? scope)
                           '((return))
                           '())))
  (cond
    [(local-lookup expr scope) => (lambda (p) (value `(local-get ,p)))]
    [(symbol? expr)               (value `(global-get ,(package-global! (scope-pkg scope) expr)))]
    [(atom? expr)                 (value `(value ,(package-value! (scope-pkg scope) expr)))]
    [else                         (case (car expr)
                                    [(begin)  (if (null? (cdr expr))
                                                (value `(void))
                                                (let ([nontail (scope-nontail scope)])
                                                  (fold-right (lambda (a d) (append (expr->instructions a nontail) d))
                                                              (expr->instructions (last expr) scope)
                                                              (cdr (drop-right expr 1)))))]
                                    [(if)     (let ([mid (control-point scope)]
                                                    [end (control-point scope)])
                                                (append
                                                  (expr->instructions (cadr expr) (scope-nontail scope))
                                                  `((branch ,mid))
                                                  (expr->instructions (caddr expr) scope)
                                                  `((jump ,end) (point ,mid))
                                                  (expr->instructions (cadddr expr) scope)
                                                  `((point ,end))))]
                                    [(set!)   (let ([p (local-lookup (cadr expr) scope)])
                                                (append
                                                  (expr->instructions (caddr expr) (scope-nontail scope))
                                                  (if p
                                                    `((local-set ,p))
                                                    `((global-set ,(package-global! (scope-pkg scope) (cadr expr)))))
                                                  (value `(void))))]
                                    [(lambda) `((block ,(package-block! (scope-pkg scope)
                                                                        (make-block (length (cadr expr))
                                                                                    (slots-needed (cddr expr))
                                                                                    (expr->instructions `(begin ,@(cddr expr))
                                                                                                        (scope-function scope
                                                                                                                        (cadr expr)))))))]
                                    [(let)    (append (append-map (lambda (expr i)
                                                                    (append (expr->instructions expr (scope-binding scope i))
                                                                            `((local-set ,(+ (length (scope-locals scope)) i)))))
                                                                  (map cadr (cadr expr))
                                                                  (iota (length (cadr expr))))
                                                      (expr->instructions `(begin ,@(cddr expr))
                                                                          (scope-let scope (map car (cadr expr)))))]
                                    [else     (let ([nontail (scope-nontail scope)])
                                                (append (append-map (lambda (expr)
                                                                      (expr->instructions expr nontail))
                                                                    expr)
                                                        (if (scope-tail-position? scope)
                                                          `((return-call))
                                                          (let ([k (control-point scope)])
                                                            `((call ,k) (point ,k))))))])]))

(define (insert-if-missing! pkg get set item)
  (let ([ls (get pkg)])
    (if (null? ls)
      (begin
        (set pkg (list item))
        0)
      (let next ([ls ls] [id 0])
        (cond
          [(equal? item (car ls)) id]
          [(null? (cdr ls)) (set-cdr! ls (list item)) (+ id 1)]
          [else (next (cdr ls) (+ id 1))])))))

(define (package-global! pkg name)
  (insert-if-missing! pkg package-globals package-globals-set! name))

(define (package-value! pkg value)
  (insert-if-missing! pkg package-values package-values-set! value))

(define (package-block! pkg block)
  (let ([id (length (package-blocks pkg))])
    (if (eqv? id 0)
      (package-blocks-set! pkg (list block))
      (set-cdr! (list-tail (package-blocks pkg) (- id 1)) (list block)))
    id))

(define (block-append! b instr)
  (if (null? (block-code b))
    (block-code-set! b (list instr))
    (append! (block-code b) (list instr))))

(define (block-point! b)
  (let ([id (block-pointc b)])
    (block-pointc-set! b (+ id 1))
    id))

(define (scope-nontail s)
  (if (not (scope-tail-position? s))
    s
    (make-scope (scope-pkg s)
                (scope-locals s)
                #f
                (scope-point-ref s))))

(define (local-lookup v scope)
  (and (symbol? v)
       (let ([p (member v (scope-locals scope))])
         (and p (- (length (scope-locals scope)) (length p))))))

(define (control-point scope)
  (let* ([ref (scope-point-ref scope)]
         [v (unbox ref)]) 
    (set-box! ref (+ v 1))
    v))

(define (scope-function scope args)
  (make-scope (scope-pkg scope) args #t (box 0)))

(define (scope-binding scope n)
  (make-scope (scope-pkg scope)
              (append (scope-locals scope) (make-list n #f))
              #f
              (scope-point-ref scope)))

(define (scope-let scope vars)
  (make-scope (scope-pkg scope)
              (append (map (lambda (v) (and (not (member v vars)) v)) (scope-locals scope))
                      vars)
              (scope-tail-position? scope)
              (scope-point-ref scope)))

  