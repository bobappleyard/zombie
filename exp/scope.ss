(import srfi-1)

(define (find-vars ep expr)
  (delete-duplicates!
    (let next ([expr expr] [scope '()] [ctx '()])
      (cond
        [(and (symbol? expr) (ep expr scope ctx)) (list expr)]
        [(atom? expr)                             '()]
        [else                                     (let ([ctx (cons expr ctx)])
                                                    (define (inner expr) (next expr scope ctx))
                                                    (case (car expr) 
                                                      [(lambda) (append-map! inner (cddr expr))]
                                                      [(let) (append (append-map! inner (map cadr (cadr expr)))
                                                                     (let ([scope (append (map car (cadr expr)) scope)])
                                                                       (append-map! (lambda (expr) (next expr scope ctx)) (cddr expr))))]
                                                      [else (append-map! inner expr)]))]))))

(define (free-var? v scope ctx)
  (not (member v scope)))

(define (boxed-var? v scope ctx)
  (and (not (null? ctx))
       (eqv? 'set! (caar ctx))
       (eqv? v (cadar ctx))
       (member 'lambda (map car ctx))
       (free-var? v scope ctx)))

(print "free: " (find-vars free-var? '(let ((x y) (y 2)) x)))
(print "boxed: " (find-vars boxed-var? '(let ((x 1)) (lambda () (set! x 2)))))