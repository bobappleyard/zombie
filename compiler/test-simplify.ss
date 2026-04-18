(import srfi-1)
(import srfi-69)
(load "simplify.ss")

(define gensym? #f)
(set! gensym 
      (let ([gensym gensym] [gensyms (make-hash-table)])
        (set! gensym? (lambda (sym) 
                        (and (symbol? sym) (hash-table-exists? gensyms sym))))
       (lambda args
          (let ([sym (apply gensym args)])
            (hash-table-set! gensyms sym #t)
            sym))))

(define (equivalent? a b)
  (define syms (make-hash-table))
  (let next ([a a] [b b])
    (cond
      [(and (symbol? a) (hash-table-exists? syms a)) (eqv? b (hash-table-ref syms a))]
      [(and (symbol? b) (hash-table-exists? syms b)) (eqv? a (hash-table-ref syms b))]
      [(and (gensym? a) (gensym? b)) (hash-table-set! syms a b) (hash-table-set! syms b a) #t]
      [(pair? a) (and (pair? b) (next (car a) (car b)) (next (cdr a) (cdr b)))]
      [else (equal? a b)])))

(define tests 0)
(define passed 0)
(define (test-case in out)
  (define got (expr->anf in))
  (set! tests (+ tests 1))
  (if (not (equivalent? got out))
    (print in ": expecting " out ", got " got)
    (set! passed (+ passed 1))))

(let ([v (gensym)] [w (gensym)])
  (test-case '(f a b)                      `(f a b))
  (test-case '(f (lambda (x) x))           `(f (lambda (x) x)))
  (test-case '(f (g a) b)                  `(let ((,v (g a))) (f ,v b)))
  (test-case '(f (g (h a)) b)              `(let ((,v (let ((,w (h a))) (g ,w)))) (f ,v b)))
  (test-case '(f (lambda (x) (g (h x) y))) `(f (lambda (x) (let ((,v (h x))) (g ,v y)))))
  (test-case '(f (begin 1) x)              `(f 1 x))
  (test-case '(f (begin a b) x)            `(let ((,v (begin a b))) (f ,v x)))
  (test-case '(f (let ((x 2)) x) y)        `(let ((,v (let ((x 2)) x))) (f ,v y))))

(print "tests: " tests ", passed: " passed)